const twine = require("twine-action-lib-service");

module.exports["ai.twine.core.business_hours"] = function (ctx, req) {
  ctx.speak("We're open 7 days a week from 9 AM until 5 PM.", true);
  ctx.revertUtterance();
};
