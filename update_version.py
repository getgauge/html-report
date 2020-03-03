import json
import os
import re

file_name = "plugin.json"


def update_version():
    with open(file_name, 'r') as f:
        data = json.load(f)
        new_version = re.sub('\d$', lambda x: str(
            int(x.group(0)) + 1), data["version"])
        data["version"] = new_version

    os.remove(file_name)

    with open(file_name, 'w') as f:
        d = json.dumps(data, indent=4, sort_keys=True)
        f.write("\n".join([l.rstrip() for l in d.split(os.linesep)]))

    return new_version


if __name__ == "__main__":
    version = update_version()
    print(version)
