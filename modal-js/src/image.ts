import {
  GenericResult,
  GenericResult_GenericStatus,
  ImageMetadata,
  ImageRegistryConfig,
} from "../proto/modal_proto/api";
import { client } from "./client";
import { imageBuilderVersion } from "./config";

/** A container image, used for starting sandboxes. */
export class Image {
  readonly imageId: string;

  /** @ignore */
  constructor(imageId: string) {
    this.imageId = imageId;
  }
}

export async function fromRegistryInternal(
  appId: string,
  tag: string,
  imageRegistryConfig?: ImageRegistryConfig,
): Promise<Image> {
  const resp = await client.imageGetOrCreate({
    appId,
    image: {
      dockerfileCommands: [`FROM ${tag}`],
      imageRegistryConfig,
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
  return new Image(resp.imageId);
}
