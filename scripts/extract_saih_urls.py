import re
with open("/tmp/saih_tajo_js.js") as f:
    text = f.read()
# url: "..." or url:'...'
for m in re.finditer(r"url[:\s]*['\"]([^'\"]+)['\"]", text):
    print(m.group(1))
# $.get("...")
for m in re.finditer(r"\.get\(['\"]([^'\"]+)['\"]", text):
    print(m.group(1))
# $.post("...")
for m in re.finditer(r"\.post\(['\"]([^'\"]+)['\"]", text):
    print(m.group(1))
# accAjax("...")
for m in re.finditer(r"accAjax\(['\"]([^'\"]+)['\"]", text):
    print(m.group(1))
