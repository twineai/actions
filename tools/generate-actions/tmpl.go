package main

const packageTemplateStr = `
{
  "name": "{{ .ActionName }}",
  "version": "0.0.1",
  "private": true,
  "dependencies": {
  }
}
`

const indexTemplateStr = `
module.exports["{{ .ActionName }}"] = function (ctx, req) {
  ctx.speak("Hello from {{ .ActionName }}!!");
};
`
