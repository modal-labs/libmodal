import { App, Image, Secret } from "modal";

const app = await App.lookup("libmodal-example", { createIfMissing: true });
const image = await Image.fromRegistry("alpine:3.21").dockerfileCommands([
  "RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
  "RUN curl -fsSL https://claude.ai/install.sh | bash",
  "ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
]);

const sb = await app.createSandbox(image);
console.log("Started Sandbox:", sb.sandboxId);

try {
  const repoUrl = "https://github.com/modal-labs/libmodal";
  const git = await sb.exec(["git", "clone", repoUrl, "/repo"]);
  await git.wait();
  console.log(`Cloned '${repoUrl}' into /repo.`);

  const claudeCmd = [
    "claude",
    "-p",
    "Summarize what this repository is about. Don't modify any code or files.",
  ];
  console.log("\nRunning command:");
  console.log(claudeCmd);
  const claude = await sb.exec(claudeCmd, {
    // Adding a PTY is important, since Claude requires it!
    pty: true,
    secrets: [
      await Secret.fromName("libmodal-anthropic-secret", {
        requiredKeys: ["ANTHROPIC_API_KEY"],
      }),
    ],
    workdir: "/repo",
  });
  await claude.wait();

  console.log("\nAgent stdout:\n");
  console.log(await claude.stdout.readText());

  const stderr = await claude.stderr.readText();
  if (stderr !== "") {
    console.log("Agent stderr:", stderr);
  }
} finally {
  await sb.terminate();
}
