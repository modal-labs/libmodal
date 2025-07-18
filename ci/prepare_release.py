"""Updates version and changelog to prepare for release.

```python
# patch release
python ci/release.py patch

# minor release
python ci/release.py minor
```
"""

import json
import re
from argparse import ArgumentParser
from pathlib import Path
from subprocess import run
from textwrap import dedent


def main():
    parser = ArgumentParser()
    parser.add_argument("update")

    args = parser.parse_args()

    # Make sure git is clean
    git_status = run(["git", "status", "--porcelain"], text=True, check=True, capture_output=True)
    if git_status.stdout != "":
        raise RuntimeError(f"git status is not clean:\n{git_status.stdout}")

    # Get updated go version
    git_tag = run(
        ["git", "tag", "--list", "modal-go*", "--sort=-v:refname"], check=True, text=True, capture_output=True
    )
    version_str = git_tag.stdout.splitlines()[0]
    match = re.match(r"modal-go/v(?P<major>[\d]+)\.(?P<minor>[\d]+)\.(?P<patch>[\d]+)", version_str)
    if not match:
        raise RuntimeError("Unable to parse modal-go version")
    current_go_verison = {key: int(match.group(key)) for key in ["major", "minor", "patch"]}
    current_go_verison[args.update] += 1
    new_go_version = f"v{current_go_verison['major']}.{current_go_verison['minor']}.{current_go_verison['patch']}"

    # Update and get new js version
    run(["npm", "version", args.update], check=True, text=True, cwd="modal-js")
    package_path = Path("modal-js") / "package.json"
    with package_path.open("r") as f:
        json_package = json.load(f)
        new_js_version = json_package["version"]

    # Update changelog with versions
    changelog_path = Path("CHANGELOG.md")
    changelog_content = changelog_path.read_text()
    version_header = f"modal-js/{new_js_version}, modal-go/{new_go_version}"

    new_header = dedent(f"""\
    ## Unreleased

    No unreleased changes.

    ## {version_header}""")

    new_changelog_content = changelog_content.replace("## Unreleased", new_header)
    changelog_path.write_text(new_changelog_content)

    run(["git", "add", str(changelog_path)])
    run(["git", "commit", "-m", f"Update changelog for {version_header}"])


if __name__ == "__main__":
    main()
