module.exports = function (context, params) {
    let name = params.name;
    if (context.secrets["open-sesame"] === undefined) {
        return {message: "I know nothing"}
    }
    return {message: "The password is " + context.secrets["open-sesame"].password}
};
