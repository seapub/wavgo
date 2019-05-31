var child = require('child_process').execFile;
var executablePath = "./splitwavwin";
var parameters = ["0.000036", "800", "400", "200", "HEYTICO.wav", "./output"];

child(executablePath, parameters, function (err, data) {

    if (err) {
        console.error(err);
        console.log(err)
        return;
    }
    console.log(data.toString());
});