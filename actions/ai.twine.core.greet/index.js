
const sample = require("lodash.sample");

const greetings = [
  "Hi, I'm Spoil Me Salon's automated assistant. What can I do for you?",
];

module.exports["ai.twine.core.greet"] = function (ctx, req) {
  const greeting = sample(greetings);
  ctx.speak(greeting, true);
  ctx.revertUtterance();
};
