
module.exports["ai.twine.core.send_to_human"] = function (ctx, req) {
  ctx.speak("ai.twine.core.send_to_human");
  ctx.connectToHuman();
};
