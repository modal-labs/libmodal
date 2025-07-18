"""CLI to help prepare and publish release.

To prepare release for

```bash
python ci/release.py version patch  # or 'minor'
```

To publish run:

```bash
python ci/release.py publish
```

"""

import json
import re
from argparse import ArgumentParser
from pathlib import Path
from subprocess import run
from textwrap import dedent


def run_cli(cmd, *args, **kwargs):
    cmd_joined = " ".join(cmd)
    print(f"> {cmd_joined}")
    return run(cmd, *args, **kwargs)


def get_current_go_version() -> dict:
    git_tag = run_cli(
        ["git", "tag", "--list", "modal-go*", "--sort=-v:refname"], check=True, text=True, capture_output=True
    )
    version_str = git_tag.stdout.splitlines()[0]
    match = re.match(r"modal-go/v(?P<major>[\d]+)\.(?P<minor>[\d]+)\.(?P<patch>[\d]+)", version_str)
    if not match:
        raise RuntimeError("Unable to parse modal-go version")
    current_go_verison = {key: int(match.group(key)) for key in ["major", "minor", "patch"]}
    return current_go_verison


def check_unreleased_has_items(changelog_content: str):
    """Check that there are items in the Unreleased section."""

    items_in_unreleased = []
    lines = changelog_content.splitlines()
    idx = 0
    while idx < len(lines):
        if lines[idx] != "## Unreleased":
            idx += 1
            continue
        # Find lines under unreleased
        idx += 1
        while idx < len(lines):
            if lines[idx].startswith("##"):
                break
            if lines[idx] and lines[idx].startswith("-"):
                items_in_unreleased.append(lines[idx])
            idx += 1

    for item in items_in_unreleased:
        if "No unreleased changes" in item:
            raise RuntimeError("Please update 'No unreleated changes' with changelog items.")

    if not items_in_unreleased:
        raise RuntimeError("Please add changelog items under the 'Unreleased' header.")


def check_git_clean():
    """Check that git status is clean."""
    git_status = run_cli(["git", "status", "--porcelain"], text=True, check=True, capture_output=True)
    if git_status.stdout != "":
        raise RuntimeError(f"git status is not clean:\n{git_status.stdout}")


def update_version(args):
    """Updates version and changelog to prepare for release.."""
    if args.update not in ["major", "minor", "patch"]:
        raise RuntimeError("update parameter must be 'major', 'minor', or 'patch'")

    # Make sure changelog has new items in "Unreleased"
    changelog_path = Path("CHANGELOG.md")
    changelog_content = changelog_path.read_text()
    check_unreleased_has_items(changelog_content)

    check_git_clean()

    # Get updated go version
    go_version = get_current_go_version()
    go_version[args.update] += 1
    new_go_version = f"v{go_version['major']}.{go_version['minor']}.{go_version['patch']}"

    # Update and get new js version
    run_cli(["npm", "version", args.update], check=True, text=True, cwd="modal-js")
    package_path = Path("modal-js") / "package.json"
    with package_path.open("r") as f:
        json_package = json.load(f)
        new_js_version = json_package["version"]

    # Update changelog with versions
    version_header = f"modal-js/v{new_js_version}, modal-go/{new_go_version}"

    new_header = dedent(f"""\
    ## Unreleased

    No unreleased changes.

    ## {version_header}""")

    new_changelog_content = changelog_content.replace("## Unreleased", new_header)
    changelog_path.write_text(new_changelog_content)

    run_cli(["git", "diff"])
    run_cli(["git", "add", str(changelog_path)])
    run_cli(["git", "commit", "-m", f"Update changelog for {version_header}"])


def publish(args):
    """Publish both modal-js and modal-go"""
    check_git_clean()
    run_cli(["npm", "publish"])

    go_version = get_current_go_version()
    go_version_str = f"v{go_version['major']}.{go_version['minor']}.{go_version['patch']}"

    run_cli(["git", "tag", f"modal-go/{go_version_str}"])
    run_cli(["git", "push", "--tags"])
    run_cli(
        ["go", "list", "-m", f"github.com/modal-labs/libmodal/modal-go@{go_version_str}"],
        env={"GOPROXY": "proxy.golang.org"},
    )


def main():
    """Entrypoint for preparing and publishing release."""
    parser = ArgumentParser()
    subparsers = parser.add_subparsers(required=True)
    version_parser = subparsers.add_parser("version")
    version_parser.add_argument("update")
    version_parser.set_defaults(func=update_version)

    publish_parser = subparsers.add_parser("publish")
    publish_parser.set_defaults(func=publish)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
