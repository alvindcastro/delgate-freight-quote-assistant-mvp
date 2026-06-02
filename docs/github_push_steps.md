# GitHub Push Steps

This repository already has a `main` branch and an `origin` remote. From the
project root, check status, commit your changes, then push:

```bash
git status
git add <changed-files>
git commit -m "Describe the change"
git pull --rebase
git push
```

Do not commit `.env` files. The `.gitignore` already excludes them.
