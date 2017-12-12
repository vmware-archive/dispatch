local BasePlugin = require "kong.plugins.base_plugin"
local cjson = require "cjson"

local ServerlessTransformerHandler = BasePlugin:extend()

function ServerlessTransformerHandler:new()
  ServerlessTransformerHandler.super.new(self, "serverless-transformer")
end

local function iter(config_array)
  return function(config_array, i, previous_name, previous_value)
    i = i + 1
    local current_pair = config_array[i]
    if current_pair == nil then -- n + 1
      return nil
    end

    local current_name, current_value = current_pair:match("^([^:]+):*(.-)$")
    if current_value == "" then
      current_value = nil
    end

    return i, current_name, current_value
  end, config_array, 0
end

local function is_json_body(header)
  local content_type = header["content-type"]
  return content_type and string.find(string.lower(content_type), "application/json", nil, true)
end

local function parse_json(body)
  if body then
    local status, res = pcall(cjson.decode, body)
    if status then
      return true, res
    else
      return false
    end
  end
end

local function transform_header(conf)
  for _, name, value in iter(conf.add.header) do
    if not ngx.req.get_headers()[name] then
      ngx.req.set_header(name, value)
      if name:lower() == "host" then -- Host header has a special treatment
        ngx.var.upstream_host = value
      end
    end
  end
end

local function transform_method(conf)
  if conf.http_method then
    local org_http_method = ngx.req.get_method()
    local rpl_http_method = conf.http_method:upper()

    ngx.req.set_method(ngx["HTTP_" .. rpl_http_method])

    -- for GET to POST transform:
    -- tranform query strings into the req body
    if org_http_method == "GET" and rpl_http_method == "POST" then
      local body = {}
      local args = ngx.req.get_uri_args()
      for key, val in pairs(args) do
        -- TODO: test if multi-value querystring works
        -- e.g. ?hello=world&hello=vmware
        body[key] = val
      end
      return body
    end
  end
  return nil
end

local function substitute_payload(conf, result)

  -- read body
  ngx.req.read_body()
  local data = ngx.req.get_body_data()
  if data then
    ngx.log(ngx.DEBUG, "request body: " .. data)
  else
    data = "{}"
  end

  local ok, json_data = parse_json(data)
  if ok then
    ngx.log(ngx.DEBUG, "request body is json")
    result[conf.substitute.input] = json_data
  else
    ngx.log(ngx.DEBUG, "request body is not json")
    result[conf.substitute.input] = data
  end
  ngx.log(ngx.DEBUG, "after substitute: payload: " .. cjson.encode(result))
  return result
end

-- insert specific headers into request body
local function insert_header_to_payload(conf, result)

  local function get_bool_from_string(str)
    if str == "true" then
      return true
    elseif str == "false" then
      return false
    end
    return str
  end

  -- insert header to the payload
  if conf.insert_to_body.header then
    for _, name, default_value in iter(conf.insert_to_body.header) do
      local prefixed_name = conf.header_prefix_for_insertion .. name

      ngx.log(ngx.DEBUG, "insert header [".. prefixed_name .. "] to payload")
      local value = ngx.req.get_headers()[prefixed_name]
      if value  == nil then
        ngx.log(ngx.DEBUG, "get empty value from req header")
        value = default_value
      else
         ngx.log(ngx.DEBUG, "get value [" .. value .. "] from req header")
      end
      result[name] = get_bool_from_string(value)
    end
  end
  ngx.log(ngx.DEBUG, "after insert_header: payload: " .. cjson.encode(result))
  return result
end

local function tranform_request(conf)

  transform_header(conf)

  local result = {}
  local args = transform_method(conf)
  if args then
    -- use the GET req querystrings as input
    result[conf.substitute.input] = args
    -- hack: lua-nginx-module requires to read
    --       the body before calling the ngx.req.set_body_data
    ngx.req.read_body()
  else
    -- move the original payload into the "input" envolope
    result = substitute_payload(conf, result)
  end

  -- insert special prefixed headers into payload
  result = insert_header_to_payload(conf, result)

  -- set the transformed data to payload
  result = cjson.encode(result)
  ngx.req.set_body_data(result)
  -- update content-length and content-type
  ngx.req.set_header("content-length", #result)
  ngx.req.set_header("content-type", "application/json")
end

function ServerlessTransformerHandler:access(conf)
  ServerlessTransformerHandler.super.access(self)
  ngx.log(ngx.DEBUG, "serverless transfer")

  if conf.enable.input then
    tranform_request(conf)
  end
end

----------------------------------------------------------
-- transform response
----------------------------------------------------------

function ServerlessTransformerHandler:header_filter(conf)
  ServerlessTransformerHandler.super.header_filter(self)

  local ctx = ngx.ctx
  ctx.rt_body_chunks = {}
  ctx.rt_body_chunk_number = 1

  -- make changes only if the response is json
  if conf.enable.output and is_json_body(ngx.header) then
    -- clear content-length header as the body content changed
    ngx.header["content-length"] = nil
  end
end

local function substitute_json_response(field, data)
  local ok = false
  ok, data = parse_json(data)
  if not ok then
      return false
  end
  local result = data
  if data[field] then
    result = data[field]
  end
  return true, cjson.encode(result)
end

local function substitute_response(conf)
  local ctx = ngx.ctx
  local chunk, eof = ngx.arg[1], ngx.arg[2]
  if eof then
    local data = table.concat(ctx.rt_body_chunks)
    ngx.log(ngx.DEBUG, "response body:" .. data)

    local ok = false
    ok, data = substitute_json_response(conf.substitute.output, data)
    if ok then
      ngx.log(ngx.DEBUG, "response body after transform:" .. data)
      ngx.log(ngx.DEBUG, "response transform done")
      ngx.arg[1] = data
    else
      ngx.log(ngx.ERR, "response transform error")
      ngx.arg[1] = nil
    end
  else
    ctx.rt_body_chunks[ctx.rt_body_chunk_number] = chunk
    ctx.rt_body_chunk_number = ctx.rt_body_chunk_number + 1
    ngx.arg[1] = nil
  end
end

function ServerlessTransformerHandler:body_filter(conf)
  ServerlessTransformerHandler.super.body_filter(self)

  -- make changes only if the response is json
  if conf.enable.output and is_json_body(ngx.header) then
    substitute_response(conf)
  end
end

ServerlessTransformerHandler.PRIORITY = 800
ServerlessTransformerHandler.VERSION = "0.0.1"
return ServerlessTransformerHandler
