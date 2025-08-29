import {
  GenericResult,
  GenericResult_GenericStatus,
  ImageMetadata,
  RegistryAuthType,
  ImageRegistryConfig,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { App } from "./app";
import { Secret } from "./secret";
import { imageBuilderVersion } from "./config";
import { ClientError } from "nice-grpc";
import { Status } from "nice-grpc";
import { NotFoundError } from "./errors";

/** Options for deleting an Image. */
export type ImageDeleteOptions = Record<never, never>;

/** A container image, used for starting Sandboxes. */
export class Image {
  #imageId: string;
  #tag: string;
  #imageRegistryConfig?: ImageRegistryConfig;

  /** @ignore */
  constructor(
    imageId: string,
    tag: string,
    imageRegistryConfig?: ImageRegistryConfig,
  ) {
    this.#imageId = imageId;
    this.#tag = tag;
    this.#imageRegistryConfig = imageRegistryConfig;
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

    const resp = await client.imageGetOrCreate({
      appId: app.appId,
      image: {
        dockerfileCommands: [`FROM ${this.#tag}`],
        imageRegistryConfig: this.#imageRegistryConfig,
      },
      builderVersion: imageBuilderVersion(),
    });

    let result: GenericResult;
    let metadata: ImageMetadata | undefined = undefined;

    if (resp.result?.status) {
      // Image has already been built
      result = resp.result;
      metadata = resp.metadata;
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
            metadata = item.metadata;
            break;
          }
          // Ignore all log lines and progress updates.
        }
      }
      result = resultJoined;
    }

    void metadata; // Note: Currently unused.

    if (result.status === GenericResult_GenericStatus.GENERIC_STATUS_FAILURE) {
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
    this.#imageId = resp.imageId;
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
