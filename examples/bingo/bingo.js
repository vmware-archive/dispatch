"use strict";

const request = require('request');

const buzzwords = ["cloud", "native", "serverless", "platform", "infrastructure", "integration", "framework"];

function bsCount(text) {
    let ws = text.toLowerCase().split(/\W+/);
    return buzzwords.filter(b => ws.includes(b)).length
}

function sayBingo(channel, oauthToken) {
    const options = {
        headers: {
            "Authorization": `Bearer ${oauthToken}`
        },
        json: {
            channel: channel,
            text: "BINGO! :wink:",
        }
    };
    request.post('https://slack.com/api/chat.postMessage', options, (error, response, body) => {
        if (!error && response.statusCode === 200) {
            console.log("BINGO!");
        }
    });
}

module.exports = (context, payload) => {
    if (payload.token !== context.secrets.verificationToken) {
        return {error: "TokenCheckFailed"};
    }
    if (payload.challenge && payload.type === "url_verification") {
        return {challenge: payload.challenge};
    }

    if (payload.event && payload.event.type === "message") {
        if (bsCount(payload.event.text) >= 3) {
            sayBingo(payload.event.channel, context.secrets.oauthToken);
        }
    }
    return {};
};
