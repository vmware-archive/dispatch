module.exports = function (params) {
    let name = params.name;
    if (params._meta.secrets === undefined) {
        return {message: "I know nothing"}
    }
    return {message: "The password is " + params._meta.secrets["open-sesame"].password}
};
