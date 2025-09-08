import {
  GenericResult,
  GenericResult_GenericStatus,
  RegistryAuthType,
  ImageRegistryConfig,
  Image as ImageProto,
  GPUConfig,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { App, parseGpuConfig } from "./app";
import { Secret } from "./secret";
import { imageBuilderVersion } from "./config";
import { ClientError } from "nice-grpc";
import { Status } from "nice-grpc";
import { NotFoundError, InvalidError } from "./errors";

/** Options for deleting an Image. */
export type ImageDeleteOptions = Record<never, never>;

/** Options for Image.dockerfileCommands(). */
export type ImageDockerfileCommandsOptions = {
  /** Secrets that will be made available to this layer's build environment. */
  secrets?: Secret[];

  /** GPU reservation for this layer's build environment (e.g. "A100", "T4:2", "A100-80GB:4"). */
  gpu?: string;

  /** Ignore cached builds for this layer, similar to 'docker build --no-cache'. */
  forceBuild?: boolean;
};

/** Represents a single image layer with its build configuration. */
type Layer = {
  commands: string[];
  secrets?: Secret[];
  gpuConfig?: GPUConfig;
  forceBuild?: boolean;
};

/** A container image, used for starting Sandboxes. */
export class Image {
  #imageId: string;
  #tag: string;
  #imageRegistryConfig?: ImageRegistryConfig;
  #layers: Layer[];

  /** @ignore */
  constructor(
    imageId: string,
    tag: string,
    imageRegistryConfig?: ImageRegistryConfig,
    layers?: Layer[],
  ) {
    this.#imageId = imageId;
    this.#tag = tag;
    this.#imageRegistryConfig = imageRegistryConfig;
    this.#layers = layers || [
      {
        commands: [],
        secrets: undefined,
        gpuConfig: undefined,
        forceBuild: false,
      },
    ];
  }
  get imageId(): string {
    return this.#imageId;
  }

  /**
   * Creates an Image from an Image ID
   *
   * @param imageId - Image ID.
   */
  static async fromId(imageId: string): Promise<Image> {
    try {
      const resp = await client.imageFromId({ imageId });
      return new Image(resp.imageId, "");
    } catch (err) {
      if (err instanceof ClientError && err.code === Status.NOT_FOUND)
        throw new NotFoundError(err.details);
      if (
        err instanceof ClientError &&
        err.code === Status.FAILED_PRECONDITION &&
        err.details.includes("Could not find image with ID")
      )
        throw new NotFoundError(err.details);
      throw err;
    }
  }

  /**
   * Creates an Image from a raw registry tag, optionally using a Secret for authentication.
   *
   * @param tag - The registry tag for the Image.
   * @param secret - Optional. A Secret containing credentials for registry authentication.
   */
  static fromRegistry(tag: string, secret?: Secret): Image {
    let imageRegistryConfig;
    if (secret) {
      if (!(secret instanceof Secret)) {
        throw new TypeError(
          "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
        );
      }
      imageRegistryConfig = {
        registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_STATIC_CREDS,
        secretId: secret.secretId,
      };
    }
    return new Image("", tag, imageRegistryConfig);
  }

  /**
   * Creates an Image from a raw registry tag, optionally using a Secret for authentication.
   *
   * @param tag - The registry tag for the Image.
   * @param secret - A Secret containing credentials for registry authentication.
   */
  static fromAwsEcr(tag: string, secret: Secret): Image {
    let imageRegistryConfig;
    if (secret) {
      if (!(secret instanceof Secret)) {
        throw new TypeError(
          "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
        );
      }
      imageRegistryConfig = {
        registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_AWS,
        secretId: secret.secretId,
      };
    }
    return new Image("", tag, imageRegistryConfig);
  }

  /**
   * Creates an Image from a raw registry tag, optionally using a Secret for authentication.
   *
   * @param tag - The registry tag for the Image.
   * @param secret - A Secret containing credentials for registry authentication.
   */
  static fromGcpArtifactRegistry(tag: string, secret: Secret): Image {
    let imageRegistryConfig;
    if (secret) {
      if (!(secret instanceof Secret)) {
        throw new TypeError(
          "secret must be a reference to an existing Secret, e.g. `await Secret.fromName('my_secret')`",
        );
      }
      imageRegistryConfig = {
        registryAuthType: RegistryAuthType.REGISTRY_AUTH_TYPE_GCP,
        secretId: secret.secretId,
      };
    }
    return new Image("", tag, imageRegistryConfig);
  }

  private static validateDockerfileCommands(commands: string[]): void {
    for (const command of commands) {
      const trimmed = command.trim().toUpperCase();
      if (trimmed.startsWith("COPY ") && !trimmed.startsWith("COPY --FROM=")) {
        throw new InvalidError(
          "COPY commands that copy from local context are not yet supported.",
        );
      }
    }
  }

  /**
   * Extend an image with arbitrary Dockerfile-like commands.
   *
   * Each call creates a new Image layer that will be built sequentially.
   * The provided options apply only to this layer.
   *
   * @param commands - Array of Dockerfile commands as strings
   * @param options - Optional configuration for this layer's build
   * @returns A new Image instance
   */
  dockerfileCommands(
    commands: string[],
    options?: ImageDockerfileCommandsOptions,
  ): Image {
    if (commands.length === 0) {
      return this;
    }

    Image.validateDockerfileCommands(commands);

    const newLayer: Layer = {
      commands: [...commands],
      secrets: options?.secrets,
      gpuConfig: options?.gpu ? parseGpuConfig(options.gpu) : undefined,
      forceBuild: options?.forceBuild,
    };

    return new Image("", this.#tag, this.#imageRegistryConfig, [
      ...this.#layers,
      newLayer,
    ]);
  }

  /**
   * Eagerly builds an Image on Modal.
   *
   * @param app - App to use to build the Image.
   */
  async build(app: App): Promise<Image> {
    if (this.imageId !== "") {
      // Image is already built with an Image ID
      return this;
    }

    let baseImageId: string | undefined;

    for (let i = 0; i < this.#layers.length; i++) {
      const layer = this.#layers[i];

      const secretIds = layer.secrets?.map((secret) => secret.secretId) || [];
      const gpuConfig = layer.gpuConfig;

      let dockerfileCommands: string[];
      let baseImages: Array<{ dockerTag: string; imageId: string }>;

      if (i === 0) {
        dockerfileCommands = [`FROM ${this.#tag}`, ...layer.commands];
        baseImages = [];
      } else {
        dockerfileCommands = ["FROM base", ...layer.commands];
        baseImages = [{ dockerTag: "base", imageId: baseImageId! }];
      }

      const resp = await client.imageGetOrCreate({
        appId: app.appId,
        image: ImageProto.create({
          dockerfileCommands,
          imageRegistryConfig: this.#imageRegistryConfig,
          secretIds,
          gpuConfig,
          contextFiles: [],
          baseImages,
        }),
        builderVersion: imageBuilderVersion(),
        forceBuild: layer.forceBuild || false,
      });

      let result: GenericResult;

      if (resp.result?.status) {
        // Image has already been built
        result = resp.result;
      } else {
        // Not built or in the process of building - wait for build
        let lastEntryId = "";
        let resultJoined: GenericResult | undefined = undefined;
        while (!resultJoined) {
          for await (const item of client.imageJoinStreaming({
            imageId: resp.imageId,
            timeout: 55,
            lastEntryId,
          })) {
            if (item.entryId) lastEntryId = item.entryId;
            if (item.result?.status) {
              resultJoined = item.result;
              break;
            }
            // Ignore all log lines and progress updates.
          }
        }
        result = resultJoined;
      }

      if (
        result.status === GenericResult_GenericStatus.GENERIC_STATUS_FAILURE
      ) {
        throw new Error(
          `Image build for ${resp.imageId} failed with the exception:\n${result.exception}`,
        );
      } else if (
        result.status === GenericResult_GenericStatus.GENERIC_STATUS_TERMINATED
      ) {
        throw new Error(
          `Image build for ${resp.imageId} terminated due to external shut-down. Please try again.`,
        );
      } else if (
        result.status === GenericResult_GenericStatus.GENERIC_STATUS_TIMEOUT
      ) {
        throw new Error(
          `Image build for ${resp.imageId} timed out. Please try again with a larger timeout parameter.`,
        );
      } else if (
        result.status !== GenericResult_GenericStatus.GENERIC_STATUS_SUCCESS
      ) {
        throw new Error(
          `Image build for ${resp.imageId} failed with unknown status: ${result.status}`,
        );
      }

      // the new image is the base for the next layer
      baseImageId = resp.imageId;
    }
    this.#imageId = baseImageId!;
    return this;
  }

  /** Delete an Image by ID. Warning: This removes an *entire Image*, and cannot be undone. */
  static async delete(
    imageId: string,
    _: ImageDeleteOptions = {},
  ): Promise<void> {
    const image = await Image.fromId(imageId);
    await client.imageDelete({ imageId: image.imageId });
  }
}
