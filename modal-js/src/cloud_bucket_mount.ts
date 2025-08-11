import {
  CloudBucketMount_BucketType,
  CloudBucketMount as CloudBucketMountProto,
} from "../proto/modal_proto/api";
import { Secret } from "./secret";

/** Cloud bucket mounts provide access to cloud storage buckets within Modal functions. */
export class CloudBucketMount {
  readonly bucketName: string;
  readonly secret?: Secret;
  readonly readOnly: boolean;
  readonly requesterPays: boolean;
  readonly bucketEndpointUrl?: string;
  readonly keyPrefix?: string;
  readonly oidcAuthRoleArn?: string;

  constructor(
    bucketName: string,
    options: {
      secret?: Secret;
      readOnly?: boolean;
      requesterPays?: boolean;
      bucketEndpointUrl?: string;
      keyPrefix?: string;
      oidcAuthRoleArn?: string;
    } = {},
  ) {
    this.bucketName = bucketName;
    this.secret = options.secret;
    this.readOnly = options.readOnly ?? false;
    this.requesterPays = options.requesterPays ?? false;
    this.bucketEndpointUrl = options.bucketEndpointUrl;
    this.keyPrefix = options.keyPrefix;
    this.oidcAuthRoleArn = options.oidcAuthRoleArn;

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
