Generate release notes for version: $ARGUMENTS

$ARGUMENTS must be a git tag (e.g., `v1.2.3`). If the tag does not exist yet, use
today's date for the release date.

Gather the data by running these commands:

1. Get the previous tag:

   ```
   git describe --tags --abbrev=0 HEAD
   ```

   If $ARGUMENTS already exists as a tag, use:

   ```
   git describe --tags --abbrev=0 $ARGUMENTS^
   ```

2. Get the release date:

   If $ARGUMENTS exists as a tag:

   ```
   git log -1 --format=%as $ARGUMENTS
   ```

   Otherwise, use today's date.

3. Get commit messages between the previous tag (from step 1) and the release ref:

   ```
   git log <previous_tag>..<ref> --pretty=format:"%s"
   ```

   Where `<ref>` is $ARGUMENTS if the tag exists, or `HEAD` otherwise.

4. Get merged PRs since the previous tag date:

   ```
   gh pr list --repo danroc/geoblock --state merged --limit 200 \
     --search "merged:>=<previous_tag_date>" \
     --json number,title,author,url,mergedAt
   ```

   Where `<previous_tag_date>` is the date of the previous tag:

   ```
   git log -1 --format=%as <previous_tag>
   ```

   Keep only PRs whose number (as `#<number>`) appears in the commit messages from
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
