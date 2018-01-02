'use strict';

const request = require('request');

module.exports = function (context, params) {
    let completeWebhook = context.secrets["slack_url"];
    let vmName = params.metadata.vm_name;
    let message = params.message;
    let vmId = params.metadata.vm_id;
    let sourceTmpl = params.metadata.src_template;
    let text = `\n${message}\n_VM Name_:\t*${vmName}*\n_Vm ID_:\t*${vmId}*\n_Source template_:\t*${sourceTmpl}*\n`;
    request.post(
        completeWebhook,
        {
            json: {
                text: text
            }
        },
        (err, response, body) => {
            if (!err) {
                return {
                    message: 'Slack message sent'
                };
            } else {
                return {
                    error: err
                };
            }
        });
    return {text: text, url: completeWebhook};
};
