module.exports = function (context, params) {
    let name = "Noone";
    if (params.name) {
        name = params.name;
    }
    let place = "Nowhere";
    if (params.place) {
        place = params.place;
    }
    console.log(`returning message name=${name} and place=${place}`)
    return { myField: 'Hello, ' + name + ' from ' + place }
};
