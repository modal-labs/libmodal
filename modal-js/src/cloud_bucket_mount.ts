import {
  CloudBucketMount_BucketType,
  CloudBucketMount as CloudBucketMountProto,
} from "../proto/modal_proto/api";
import { Secret } from "./secret";

/** Cloud bucket mounts provide access to cloud storage buckets within Modal functions. */
export class CloudBucketMount {
  readonly bucketName: string;
  readonly bucketType: CloudBucketMount_BucketType;
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

    // Determine bucket type from endpoint URL
    if (this.bucketEndpointUrl) {
      const url = new URL(this.bucketEndpointUrl);
      if (url.hostname.endsWith("r2.cloudflarestorage.com")) {
        this.bucketType = CloudBucketMount_BucketType.R2;
      } else if (url.hostname.endsWith("storage.googleapis.com")) {
        this.bucketType = CloudBucketMount_BucketType.GCP;
      } else {
        console.warn(
          "CloudBucketMount received unrecognized bucket endpoint URL. " +
            "Assuming AWS S3 configuration as fallback.",
        );
        this.bucketType = CloudBucketMount_BucketType.S3;
      }
    } else {
      // Just assume S3; this is backwards and forwards compatible
      this.bucketType = CloudBucketMount_BucketType.S3;
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

  /** Convert this CloudBucketMount to a protobuf message. */
  toProto(mountPath: string): CloudBucketMountProto {
    return {
      bucketName: this.bucketName,
      mountPath,
      credentialsSecretId: this.secret?.secretId ?? "",
      readOnly: this.readOnly,
      bucketType: this.bucketType,
      requesterPays: this.requesterPays,
      bucketEndpointUrl: this.bucketEndpointUrl,
      keyPrefix: this.keyPrefix,
      oidcAuthRoleArn: this.oidcAuthRoleArn,
    };
  }
}
