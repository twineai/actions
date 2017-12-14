
const sample = require("lodash.sample");

const greetings = [
  "Hello, and thank you for calling! How may I help you today?",
  "Hi, I'm Spoil Me Salon's automated assistant. What can I do for you?",
];

module.exports["ai.twine.core.greet"] = function (ctx, req) {
  const greeting = sample(greetings);
  ctx.speak(greeting, true);
};
