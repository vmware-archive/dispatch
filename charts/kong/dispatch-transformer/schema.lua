-- entries must have colons to set the key and value apart
local function check_for_value(value)
  for i, entry in ipairs(value) do
    local ok = string.find(entry, ":")
    if not ok then
      return false, "key '" .. entry .. "' has no value"
    end
  end
  return true
end

local function check_method(value)
  if not value then
    return true
  end
  local method = value:upper()
  local ngx_method = ngx["HTTP_" .. method]
  if not ngx_method then
    return false, method .. " is not supported"
  end
  return true
end

return {
  fields = {
    http_method = {type = "string", default = "POST", func = check_method},
    header_prefix_for_insertion = {type = "string", default = "x-dispatch-"},
    add = {
      type = "table",
      schema = {
        fields = {
          header = {type = "array", default = {}, func = check_for_value},
          -- non-overidable internal headers
          internal_header = {type = "array", default = {}, func = check_for_value},
        }
      }
    },
    append = {
      type = "table",
      schema = {
        fields = {
          querystring = {type = "array", default = {}, func = check_for_value}
        }
      }
    },
    insert_to_body = {
      type = "table",
      schema = {
        fields = {
          -- insert headers into request body
          header = {type = "array", default = {}, func = check_for_value},
        }
      }
    },
    substitute = {
      type = "table",
      schema = {
        fields = {
          input  = { type = "string", default = "input" },
          output = { type = "string", default = "output" },
          http_context = { type = "string", default = "httpContext" },
        }
      }
    },
    enable = {
      type = "table",
      schema = {
        fields = {
          input  = { type = "boolean", default = true },
          output = { type = "boolean", default = true },
          http_context = { type = "boolean", default = true },
        }
      }
    }
  }
}
