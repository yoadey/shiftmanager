# GitHub Actions Workflows

These two workflow files (`ci.yml`, `deploy.yml`) belong under
`.github/workflows/` to be picked up by GitHub Actions.

They live here temporarily because the Personal Access Token used for the
initial push lacked the `workflow` scope, which GitHub requires to create or
update files under `.github/workflows/`.

## To activate CI/CD

Move them into place and commit with a token that has the `workflow` scope
(or do it via the GitHub web UI):

```bash
mkdir -p .github/workflows
git mv deploy/github-workflows/ci.yml     .github/workflows/ci.yml
git mv deploy/github-workflows/deploy.yml .github/workflows/deploy.yml
git commit -m "ci: enable GitHub Actions workflows"
git push
```

- `ci.yml` — lint + test backend, typecheck + build frontend, build & push Docker image
- `deploy.yml` — deploy to Kubernetes on push to `main`
