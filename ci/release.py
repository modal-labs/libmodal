"""CLI to help prepare and publish release.

To prepare stable release:

```bash
python ci/release.py version patch  # or 'minor' or 'major'
```

To prepare dev release:

```bash
python ci/release.py version patch --dev
```

To publish stable release:

```bash
python ci/release.py publish
```

To publish dev release:

```bash
python ci/release.py publish --dev
```

"""

import json
from argparse import ArgumentParser
from pathlib import Path
from subprocess import run
from textwrap import dedent


def run_cli(cmd, *args, **kwargs):
    cmd_joined = " ".join(cmd)
    print(f"> {cmd_joined}")
    return run(cmd, *args, **kwargs, check=True)


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
    git_status = run_cli(["git", "status", "--porcelain"], text=True, capture_output=True)
    if git_status.stdout != "":
        raise RuntimeError(f"git status is not clean:\n{git_status.stdout}")


def get_current_js_version():
    package_path = Path("modal-js") / "package.json"
    with package_path.open("r") as f:
        json_package = json.load(f)
        return json_package["version"]


def update_version(args):
    """Updates version and changelog and prepare a release PR."""
    if args.update not in ["major", "minor", "patch"]:
        raise RuntimeError("update parameter must be 'major', 'minor', or 'patch'")

    check_git_clean()

    if args.dev:
        current_version = get_current_js_version()

        if "-dev." in current_version:
            run_cli(["npm", "version", "prerelease", "--no-git-tag-version"], cwd="modal-js")
        else:
            run_cli(["npm", "version", f"pre{args.update}", "--preid=dev", "--no-git-tag-version"], cwd="modal-js")

        new_version = get_current_js_version()

        run_cli(["git", "diff"])

        commit_message = f"[DEV-RELEASE] Prepare dev release for modal-js/v{new_version}, modal-go/v{new_version}"
        if args.dry_run:
            print("\nDRY RUN: Would create commit with message:")
            print(commit_message)
            run_cli(["git", "restore", "--", "modal-js/package.json", "modal-js/package-lock.json"])
        else:
            run_cli(["git", "add", "modal-js/package.json", "modal-js/package-lock.json"])
            run_cli(["git", "commit", "-m", commit_message])
    else:
        changelog_path = Path("CHANGELOG.md")
        changelog_content = changelog_path.read_text()
        check_unreleased_has_items(changelog_content)

        run_cli(["npm", "version", args.update, "--no-git-tag-version"], text=True, cwd="modal-js")
        new_version = get_current_js_version()

        version_header = f"modal-js/v{new_version}, modal-go/v{new_version}"

        new_header = dedent(f"""\
        ## Unreleased

        No unreleased changes.

        ## {version_header}""")

        new_changelog_content = changelog_content.replace("## Unreleased", new_header)
        changelog_path.write_text(new_changelog_content)

        run_cli(["git", "diff"])
        run_cli(["git", "add", "modal-js/package.json", "modal-js/package-lock.json", str(changelog_path)])

        commit_message = f"[RELEASE] Prepare release for {version_header}"
        if args.dry_run:
            print("\nDRY RUN: Would create commit with message:")
            print(commit_message)
            run_cli(["git", "reset", "HEAD"])
            run_cli(
                ["git", "restore", "--", "modal-js/package.json", "modal-js/package-lock.json", str(changelog_path)]
            )
        else:
            run_cli(["git", "commit", "-m", commit_message])


def publish(args):
    """Publish both modal-js and modal-go"""
    check_git_clean()

    version = get_current_js_version()
    js_tag = f"modal-js/v{version}"
    go_tag = f"modal-go/v{version}"

    if args.dry_run:
        print("\nDRY RUN: Would execute the following operations:")
        print("  - git push (push version commit)")
        if args.dev:
            print("  - npm publish --tag next (in modal-js/)")
        else:
            print("  - npm publish (in modal-js/)")
        print(f"  - Create and push git tags: {js_tag}, {go_tag}")
        return

    run_cli(["git", "push"])
    run_cli(["git", "tag", js_tag])
    run_cli(["git", "tag", go_tag])
    run_cli(["git", "push", "--tags"])

    if args.dev:
        run_cli(["npm", "publish", "--tag", "next"], cwd="modal-js")
    else:
        run_cli(["npm", "publish"], cwd="modal-js")

    run_cli(["curl", f"https://proxy.golang.org/github.com/modal-labs/libmodal/modal-go/@v/v{version}.info"])


def main():
    """Entrypoint for preparing and publishing release."""
    parser = ArgumentParser()
    subparsers = parser.add_subparsers(required=True)

    version_parser = subparsers.add_parser("version")
    version_parser.add_argument("update")
    version_parser.add_argument("--dev", action="store_true", help="Create dev release")
    version_parser.add_argument("--dry-run", action="store_true", help="Show what would be done without making changes")
    version_parser.set_defaults(func=update_version)

    publish_parser = subparsers.add_parser("publish")
    publish_parser.add_argument("--dev", action="store_true", help="Publish dev release")
    publish_parser.add_argument("--dry-run", action="store_true", help="Show what would be done without making changes")
    publish_parser.set_defaults(func=publish)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
