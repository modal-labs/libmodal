import {
  CloudBucketMount_BucketType,
  CloudBucketMount as CloudBucketMountProto,
} from "../proto/modal_proto/api";
import { Secret } from "./secret";
import { getDefaultClient, type ModalClient } from "./client";
import { InvalidError } from "./errors";

/** Optional parameters for {@link CloudBucketMountService#new client.cloudBucketMounts.new()}. */
export type CloudBucketMountParams = {
  secret?: Secret;
  readOnly?: boolean;
  requesterPays?: boolean;
  bucketEndpointUrl?: string;
  keyPrefix?: string;
  oidcAuthRoleArn?: string;
};

/**
 * Service for managing {@link CloudBucketMount CloudBucketMounts}.
 *
 * Normally only ever accessed via the client as:
 * ```typescript
 * const modal = new ModalClient();
 * const mount = modal.cloudBucketMounts.new("my-bucket", { readOnly: true });
 * ```
 */
export class CloudBucketMountService {
  readonly #client: ModalClient;
  constructor(client: ModalClient) {
    this.#client = client;
  }

  /** Create a new {@link CloudBucketMount}. */
  new(bucketName: string, params?: CloudBucketMountParams): CloudBucketMount {
    // Validate before construction
    if (params?.bucketEndpointUrl) {
      const url = new URL(params.bucketEndpointUrl);
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

    if (params?.requesterPays && !params.secret) {
      throw new InvalidError(
        "Credentials required in order to use Requester Pays.",
      );
    }

    if (params?.keyPrefix && !params.keyPrefix.endsWith("/")) {
      throw new InvalidError(
        "keyPrefix will be prefixed to all object paths, so it must end in a '/'",
      );
    }

    return new CloudBucketMount(
      bucketName,
      params?.secret,
      params?.readOnly ?? false,
      params?.requesterPays ?? false,
      params?.bucketEndpointUrl,
      params?.keyPrefix,
      params?.oidcAuthRoleArn,
    );
  }
}

/** Cloud Bucket Mounts provide access to cloud storage buckets within Modal Functions. */
export class CloudBucketMount {
  readonly bucketName: string;
  readonly secret?: Secret;
  readonly readOnly: boolean;
  readonly requesterPays: boolean;
  readonly bucketEndpointUrl?: string;
  readonly keyPrefix?: string;
  readonly oidcAuthRoleArn?: string;

  /**
   * @deprecated Use {@link CloudBucketMountService#new client.cloudBucketMounts.new()} instead.
   */
  constructor(
    bucketName: string,
    paramsOrSecret?: CloudBucketMountParams | Secret,
    readOnly?: boolean,
    requesterPays?: boolean,
    bucketEndpointUrl?: string,
    keyPrefix?: string,
    oidcAuthRoleArn?: string,
  ) {
    this.bucketName = bucketName;

    // Support both old object-style params and new individual parameters
    if (
      paramsOrSecret &&
      typeof paramsOrSecret === "object" &&
      "secretId" in paramsOrSecret
    ) {
      // New style: individual parameters (used internally by service)
      this.secret = paramsOrSecret as Secret;
      this.readOnly = readOnly ?? false;
      this.requesterPays = requesterPays ?? false;
      this.bucketEndpointUrl = bucketEndpointUrl;
      this.keyPrefix = keyPrefix;
      this.oidcAuthRoleArn = oidcAuthRoleArn;
    } else {
      // Old style: params object (for backward compatibility)
      const params = (paramsOrSecret as CloudBucketMountParams) || {};
      this.secret = params.secret;
      this.readOnly = params.readOnly ?? false;
      this.requesterPays = params.requesterPays ?? false;
      this.bucketEndpointUrl = params.bucketEndpointUrl;
      this.keyPrefix = params.keyPrefix;
      this.oidcAuthRoleArn = params.oidcAuthRoleArn;
    }
  }

  /**
   * @deprecated Use {@link CloudBucketMountService#new client.cloudBucketMounts.new()} instead.
   */
  static new(
    bucketName: string,
    params?: CloudBucketMountParams,
  ): CloudBucketMount {
    return getDefaultClient().cloudBucketMounts.new(bucketName, params);
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
