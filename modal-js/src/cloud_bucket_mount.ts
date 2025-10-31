import {
  CloudBucketMount_BucketType,
  CloudBucketMount as CloudBucketMountProto,
} from "../proto/modal_proto/api";
import type { ModalClient } from "./client";
import { Secret } from "./secret";

export type CloudBucketMountParams = {
  secret?: Secret;
  readOnly?: boolean;
  requesterPays?: boolean;
  bucketEndpointUrl?: string;
  keyPrefix?: string;
  oidcAuthRoleArn?: string;
};

/**
 * Cloud bucket mounts let functions access objects from external storage.
 *
 * @deprecated Use {@link CloudBucketMountService#new client.cloudBucketMounts.new()} instead.
 */
export class CloudBucketMount {
  readonly bucketName: string;
  readonly secret?: Secret;
  readonly readOnly: boolean;
  readonly requesterPays: boolean;
  readonly bucketEndpointUrl?: string;
  readonly keyPrefix?: string;
  readonly oidcAuthRoleArn?: string;

  constructor(bucketName: string, params: CloudBucketMountParams = {}) {
    this.bucketName = bucketName;
    this.secret = params.secret;
    this.readOnly = params.readOnly ?? false;
    this.requesterPays = params.requesterPays ?? false;
    this.bucketEndpointUrl = params.bucketEndpointUrl;
    this.keyPrefix = params.keyPrefix;
    this.oidcAuthRoleArn = params.oidcAuthRoleArn;

    if (this.bucketEndpointUrl) {
      const url = new URL(this.bucketEndpointUrl);
      if (
        !url.hostname.endsWith("r2.cloudflarestorage.com") &&
        !url.hostname.endsWith("storage.googleapis.com")
      ) {
        console.warn(
          "CloudBucketMount received unrecognized bucket endpoint URL. " +
            "Assuming AWS S3 configuration as fallback.",
        );
      }
    }

    if (this.requesterPays && !this.secret) {
      throw new Error("Credentials required in order to use Requester Pays.");
    }

    if (this.keyPrefix && !this.keyPrefix.endsWith("/")) {
      throw new Error(
        "keyPrefix will be prefixed to all object paths, so it must end in a '/'",
      );
    }
  }
}

export class CloudBucketMountService {
  readonly #client: ModalClient;

  constructor(client: ModalClient) {
    this.#client = client;
  }

  new(bucketName: string, params: CloudBucketMountParams = {}): CloudBucketMount {
    // Touch the client instance to avoid unused-private-member lint errors until the service
    // needs it for control-plane interactions.
    void this.#client;
    return new CloudBucketMount(bucketName, params);
  }
}

export function endpointUrlToBucketType(
  bucketEndpointUrl?: string,
): CloudBucketMount_BucketType {
  if (!bucketEndpointUrl) {
    return CloudBucketMount_BucketType.S3;
  }

  const url = new URL(bucketEndpointUrl);
  if (url.hostname.endsWith("r2.cloudflarestorage.com")) {
    return CloudBucketMount_BucketType.R2;
  } else if (url.hostname.endsWith("storage.googleapis.com")) {
    return CloudBucketMount_BucketType.GCP;
  } else {
    return CloudBucketMount_BucketType.S3;
  }
}

export function cloudBucketMountToProto(
  mount: CloudBucketMount,
  mountPath: string,
): CloudBucketMountProto {
  return {
    bucketName: mount.bucketName,
    mountPath,
    credentialsSecretId: mount.secret?.secretId ?? "",
    readOnly: mount.readOnly,
    bucketType: endpointUrlToBucketType(mount.bucketEndpointUrl),
    requesterPays: mount.requesterPays,
    bucketEndpointUrl: mount.bucketEndpointUrl,
    keyPrefix: mount.keyPrefix,
    oidcAuthRoleArn: mount.oidcAuthRoleArn,
  };
}
