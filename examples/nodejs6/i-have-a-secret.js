module.exports = function (context, params) {
    let name = params.name;
    if (context.secrets["password"] === undefined) {
        return {message: "I know nothing"}
    }
    return {message: "The password is " + context.secrets["password"]}
};
