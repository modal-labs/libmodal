import {
  GenericResult,
  GenericResult_GenericStatus,
  ImageMetadata,
  RegistryAuthType,
  ImageRegistryConfig,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { Secret } from "./secret";
import { imageBuilderVersion } from "./config";

/** A container image, used for starting sandboxes. */
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
   * Creates an `Image` instance from a raw registry tag, optionally using a secret for authentication.
   *
   * @param tag - The registry tag for the image.
   * @param secret - Optional. A `Secret` instance containing credentials for registry authentication.
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
   * Creates an `Image` instance from a raw registry tag, optionally using a secret for authentication.
   *
   * @param tag - The registry tag for the image.
   * @param secret - A `Secret` instance containing credentials for registry authentication.
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
   * Creates an `Image` instance from a raw registry tag, optionally using a secret for authentication.
   *
   * @param tag - The registry tag for the image.
   * @param secret - A `Secret` instance containing credentials for registry authentication.
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
   * @internal
   * Build image object
   */
  async _build(appId: string): Promise<Image> {
    if (this.imageId !== "") {
      // Image is already built with an image ID
      return this;
    }

    const resp = await client.imageGetOrCreate({
      appId,
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
}
