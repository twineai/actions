const mongoose = require("mongoose");

var ServiceSchema = mongoose.Schema({
  name: String
});

module.exports = ServiceSchema;