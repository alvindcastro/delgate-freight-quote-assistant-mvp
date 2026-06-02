# GitHub Push Steps

From the project root:

```bash
git init
git add .
git commit -m "Build freight quote assistant MVP"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/delgate-freight-quote-assistant.git
git push -u origin main
```

Before pushing, replace `YOUR_USERNAME` with your GitHub username and make sure you created the empty repository on GitHub.

Do not commit `.env` files. The `.gitignore` already excludes them.
