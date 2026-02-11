# Perfecto Imperfecto - Template Manager

A simple web app to manage survey chapter templates. Edit questions, answers, and conversation prompts through an easy-to-use browser interface — no coding required.

## Setup (one time)

1. **Install Node.js** from https://nodejs.org (download the LTS version)
2. Open a terminal / command prompt
3. Navigate to this folder:
   ```
   cd path/to/perfectoimperfecto/manager
   ```
4. Install dependencies:
   ```
   npm install
   ```

## Running the Manager

```
npm start
```

Then open your browser at **http://localhost:3456**

Press `Ctrl+C` in the terminal to stop.

## Features

- **Browse chapters** — Spanish and English chapters listed in the sidebar
- **Edit questions** — Change question text, answer options, add/remove questions
- **Supports all question types** — Radio buttons, checkboxes, and text/conversation fields
- **Preview** — Open the actual survey page in a new tab to see how it looks
- **Save** — Saves changes directly to the HTML files
- **Commit & Push** — One-click button to save your changes to Git and upload them

## How to Use

1. Click a chapter in the left sidebar
2. Edit the question text, answer options, or conversation prompts
3. Click **Save Changes** when done
4. When ready to publish, click **Commit & Push** in the top bar
5. Enter a short description of what you changed and confirm
