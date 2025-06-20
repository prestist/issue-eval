# Using issue-eval in Your Repository

This guide shows how to add this tool to your GitHub repository.

### Step 1: Simple Composite Action (Recommended)

The easiest way to use issue-eval in your repository:

```yaml
# .github/workflows/issue-eval.yml
name: issue-eval

on:
  issues:
    types: [opened]

permissions:
  issues: write
  contents: read
jobs:
  process-issue:
    runs-on: ubuntu-latest
    steps:
    - name: Run issue-eval
      uses: prestist/issue-eval@main
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        google-ai-api-key: ${{ secrets.GOOGLE_AI_API_KEY }}
        issue-number: ${{ github.event.issue.number }}
        repo-owner: ${{ github.repository_owner }}
        repo-name: ${{ github.event.repository.name }}
```

### Step 2: Set up Google AI API Key

1. **Get an API key** from [Google AI Studio](https://aistudio.google.com/)
2. **Add it to your repository secrets**:
   - Go to your repo → Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `GOOGLE_AI_API_KEY`
   - Value: Your API key

### Step 3: Test It

1. **Create a test issue** in your repository
2. **Check the Actions tab** to see the workflow run
3. **Look for the AI comment** on your issue
