import { ModalClient } from "modal";

const modal = new ModalClient();

const app = await modal.apps.fromName("libmodal-example", {
  createIfMissing: true,
});

// Docker image from https://github.com/modal-labs/modal-examples/pull/1418/files
const image = modal.images.fromRegistry("ghcr.io/thomasjpfan/docker-in-sandbox:0.0.12");

const sb = await modal.sandboxes.create(app, image, {
    command: ["/start-dockerd.sh"],
    experimentalOptions: {"enable_docker": true},
})

const dockerFile = `
FROM ubuntu
RUN apt-get update
RUN apt-get install -y cowsay curl
RUN mkdir -p /usr/share/cowsay/cows/
RUN curl -o /usr/share/cowsay/cows/docker.cow https://raw.githubusercontent.com/docker/whalesay/master/docker.cow
ENTRYPOINT ["/usr/games/cowsay", "-f", "docker.cow"]
`;

const writeHandle = await sb.open("/build/Dockerfile", "w");
const encoder = new TextEncoder();
await writeHandle.write(encoder.encode(dockerFile))
await writeHandle.close();

console.log("Building docker image")
const buildProc = await sb.exec(["docker", "build", "--network=host", "-t", "whalesay", "/build"])
await buildProc.wait();

console.log("Running docker image")
const runProc = await sb.exec(["docker", "run", "--rm", "whalesay", "Hello!"])
await runProc.wait();
const runInfo = await runProc.stdout.readText();
console.log(runInfo)

await sb.terminate();
