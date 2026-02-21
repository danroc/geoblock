Generate release notes for version: $ARGUMENTS

$ARGUMENTS must be a git tag (e.g., `v1.2.3`). If the tag does not exist yet, use
today's date for the release date.

Gather the data by running these commands:

1. Get the previous tag:

   ```bash
   git describe --tags --abbrev=0 HEAD
   ```

   If $ARGUMENTS already exists as a tag, use:

   ```bash
   git describe --tags --abbrev=0 $ARGUMENTS^
   ```

2. Get the release date:

   If $ARGUMENTS exists as a tag:

   ```bash
   git log -1 --format=%as $ARGUMENTS
   ```

   Otherwise, use today's date.

3. Get commit messages between the previous tag (from step 1) and the release ref:

   ```bash
   git log <previous_tag>..<ref> --pretty=format:"%s"
   ```

   Where `<ref>` is $ARGUMENTS if the tag exists, or `HEAD` otherwise.

4. Extract PR numbers from the commit messages (step 3) and fetch their details:

   ```bash
   gh pr view <number> --repo danroc/geoblock --json number,title,author,url,mergedAt
   ```

   Run this for each PR number referenced (as `#<number>`) in the commit messages from
   step 3.

Format the output as Markdown using this template:

```markdown
## $ARGUMENTS - <release-date>

### Breaking changes

### Action required

### New

### Improvements

### Fixes

### Security
```

Filtering:

- Exclude dependency updates (`chore(deps)`, `fix(deps)`) unless security-related
- Exclude release commits (`release: X.Y.Z`)
- Exclude CI/build-only changes unless they affect the Docker image or user-facing
  behavior

Classification:

- **Breaking changes**: removes or renames existing behavior (endpoints, env vars, log
  fields, config keys)
- **Action required**: users must take a step, but nothing breaks silently (e.g.,
  default path changed with a clear error)
- **New**: adds a capability that didn't exist before
- **Improvements**: enhances existing behavior (performance, UX, observability)
- **Fixes**: corrects a bug
- **Security**: addresses a vulnerability

Rules:

- Omit sections if they have no items
- If no sections remain, output "No user-facing changes." as the body
- Each item must include: ([#NUMBER](PR_URL), @author)
- Rewrite PR titles for clarity when needed
- Audience: self-hosters / homelab operators
- Be operational and concise; no marketing language
- Output Markdown only

After generating the release notes, ask the user whether to create or update the GitHub
release for $ARGUMENTS using the generated notes. If the user agrees, run:

```bash
gh release view $ARGUMENTS --repo danroc/geoblock
```

- If the release exists, update it:

  ```bash
  gh release edit $ARGUMENTS --repo danroc/geoblock --notes "<notes>"
  ```

- If it does not exist, create it:

  ```bash
  gh release create $ARGUMENTS --repo danroc/geoblock --title "$ARGUMENTS" --notes "<notes>"
  ```
