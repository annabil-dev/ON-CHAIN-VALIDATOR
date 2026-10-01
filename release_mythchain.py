#!/usr/bin/env python3
"""Publish a tagged Mythchain prerelease from Windows using GitHub Actions.

Requirements: Python 3, Git, and GitHub CLI (gh) authenticated to GitHub.
The existing release workflow builds the Linux packages on GitHub Actions, so WSL
does not need to be opened or invoked locally.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import time
from pathlib import Path


REPO = "annabil-dev/ON-CHAIN-VALIDATOR"
ROOT = Path(__file__).resolve().parent
RELEASE_WORKFLOW = "release.yml"
TAG_PATTERN = re.compile(r"^v\d+\.\d+\.\d+-myth-phase1$")


def run(command: list[str], *, capture: bool = False, check: bool = True) -> subprocess.CompletedProcess[str]:
    print("+", subprocess.list2cmdline(command))
    result = subprocess.run(
        command,
        cwd=ROOT,
        text=True,
        capture_output=capture,
        check=False,
    )
    if check and result.returncode:
        detail = result.stderr.strip() if capture and result.stderr else ""
        raise RuntimeError(f"Command failed ({result.returncode}): {detail or command[0]}")
    return result


def output(command: list[str]) -> str:
    result = run(command, capture=True)
    return result.stdout.strip()


def assert_clean_worktree() -> None:
    status = output(["git", "status", "--porcelain"])
    if status:
        raise RuntimeError(
            "Working tree is not clean. Commit the intended release changes first:\n" + status
        )


def assert_release_tag_is_unused(tag: str) -> None:
    local = output(["git", "tag", "--list", tag])
    remote = output(["git", "ls-remote", "--tags", "origin", f"refs/tags/{tag}"])
    if local or remote:
        raise RuntimeError(f"Tag {tag} already exists locally or on origin; refusing to overwrite it.")


def wait_for_release(tag: str, commit: str, timeout_seconds: int = 600) -> str:
    deadline = time.monotonic() + timeout_seconds
    run_id: str | None = None
    while time.monotonic() < deadline:
        runs = json.loads(
            output(
                [
                    "gh",
                    "run",
                    "list",
                    "--repo",
                    REPO,
                    "--workflow",
                    RELEASE_WORKFLOW,
                    "--limit",
                    "20",
                    "--json",
                    "databaseId,headSha,headBranch,status,conclusion",
                ]
            )
        )
        matching = [
            item
            for item in runs
            if item.get("headSha") == commit and item.get("headBranch") == tag
        ]
        if matching:
            run_id = str(matching[0]["databaseId"])
            break
        time.sleep(5)

    if not run_id:
        raise RuntimeError(
            f"No release workflow appeared for {tag}. Check Actions: "
            f"https://github.com/{REPO}/actions"
        )

    run(["gh", "run", "watch", run_id, "--repo", REPO, "--exit-status"])
    release = json.loads(
        output(
            [
                "gh",
                "release",
                "view",
                tag,
                "--repo",
                REPO,
                "--json",
                "isDraft,isPrerelease,url,assets",
            ]
        )
    )
    if release.get("isDraft") or not release.get("isPrerelease"):
        raise RuntimeError(f"GitHub created {tag}, but it is not a published prerelease.")

    asset_names = {asset["name"] for asset in release.get("assets", [])}
    version = tag.removeprefix("v")
    required = {
        f"mythprotocold-{version}-linux-amd64.tar.gz",
        f"mythprotocold-{version}-linux-arm64.tar.gz",
        f"mythprotocold_{version}_amd64.deb",
        f"mythprotocold_{version}_amd64.deb.sha256",
        f"mythprotocold_{version}_arm64.deb",
        f"mythprotocold_{version}_arm64.deb.sha256",
        "SHA256SUMS",
    }
    missing = sorted(required - asset_names)
    if missing:
        raise RuntimeError("Release is missing expected assets: " + ", ".join(missing))
    return release["url"]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("tag", help="Release tag, e.g. v0.1.2-myth-phase1")
    parser.add_argument("--yes", action="store_true", help="Skip the release confirmation prompt")
    args = parser.parse_args()

    if not TAG_PATTERN.fullmatch(args.tag):
        parser.error("tag must look like v0.1.2-myth-phase1")

    try:
        assert_clean_worktree()
        run(["gh", "auth", "status"])
        repo = output(["gh", "repo", "view", REPO, "--json", "nameWithOwner", "-q", ".nameWithOwner"])
        if repo.lower() != REPO.lower():
            raise RuntimeError(f"Expected GitHub repository {REPO}, got {repo}.")

        branch = output(["git", "branch", "--show-current"])
        if not branch:
            raise RuntimeError("A named branch is required; detached HEAD is not supported.")
        assert_release_tag_is_unused(args.tag)

        commit = output(["git", "rev-parse", "HEAD"])
        print(f"\nRelease: {args.tag}\nBranch:  {branch}\nCommit:  {commit}\n")
        if not args.yes and input(f"Push {branch} and tag {args.tag}, then publish the prerelease? [y/N] ").strip().lower() != "y":
            print("Cancelled; nothing was pushed.")
            return 0

        run(["git", "push", "origin", branch])
        run(["git", "tag", "-a", args.tag, "-m", f"Release {args.tag}"])
        try:
            run(["git", "push", "origin", args.tag])
        except Exception:
            run(["git", "tag", "-d", args.tag], check=False)
            raise

        url = wait_for_release(args.tag, commit)
        print(f"\nRelease published: {url}")
        return 0
    except (OSError, RuntimeError, subprocess.CalledProcessError) as exc:
        print(f"Release failed: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
