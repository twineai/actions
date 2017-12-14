const mongoose = require("mongoose");

var ClientSchema = mongoose.Schema({
  name: String
});

var ServiceSchema = mongoose.Schema({
  name: String
});


module.exports = ClientSchema;