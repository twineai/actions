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
module.exports["{{ .ActionName }}"] = function () {
  return "Hello from {{ .ActionName }}!!";
};
`
