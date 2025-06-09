import { readFile } from "node:fs/promises";
import path from "node:path";
import { homedir } from "node:os";
import { parse as parseToml } from "smol-toml";

/** Raw representation of the .modal.toml file. */
interface Config {
  [profile: string]: {
    server_url?: string;
    token_id?: string;
    token_secret?: string;
    environment?: string;
    imageBuilderVersion?: string;
    active?: boolean;
  };
}

/** Resolved configuration object from `Config` and environment variables. */
export interface Profile {
  serverUrl: string;
  tokenId: string;
  tokenSecret: string;
  environment?: string;
  imageBuilderVersion?: string;
}

async function readConfigFile(): Promise<Config> {
  try {
    const configContent = await readFile(path.join(homedir(), ".modal.toml"), {
      encoding: "utf-8",
    });
    return parseToml(configContent) as Config;
  } catch (err: any) {
    if (err.code === "ENOENT") {
      return {} as Config;
    }
    throw new Error(`Failed to read or parse .modal.toml: ${err.message}`);
  }
}

// Top-level await on module startup.
const config: Config = await readConfigFile();

export function getProfile(profileName?: string): Profile {
  profileName = profileName || process.env["MODAL_PROFILE"] || undefined;
  if (!profileName) {
    for (const [name, profileData] of Object.entries(config)) {
      if (profileData.active) {
        profileName = name;
        break;
      }
    }
  }
  if (profileName && !Object.hasOwn(config, profileName)) {
    throw new Error(
      `Profile "${profileName}" not found in .modal.toml. Please set the MODAL_PROFILE environment variable or specify a valid profile.`,
    );
  }
  const profileData = profileName ? config[profileName] : {};

  const profile: Partial<Profile> = {
    serverUrl:
      process.env["MODAL_SERVER_URL"] ||
      profileData.server_url ||
      "https://api.modal.com:443",
    tokenId: process.env["MODAL_TOKEN_ID"] || profileData.token_id,
    tokenSecret: process.env["MODAL_TOKEN_SECRET"] || profileData.token_secret,
    environment: process.env["MODAL_ENVIRONMENT"] || profileData.environment,
    imageBuilderVersion:
      process.env["MODAL_IMAGE_BUILDER_VERSION"] ||
      profileData.imageBuilderVersion,
  };
  if (!profile.tokenId || !profile.tokenSecret) {
    throw new Error(
      `Profile "${profileName}" is missing token_id or token_secret. Please set them in .modal.toml or as environment variables.`,
    );
  }
  return profile as Profile; // safe to null-cast because of check above
}

export const profile = getProfile(process.env["MODAL_PROFILE"] || undefined);

export function environmentName(environment?: string): string {
  return environment || profile.environment || "";
}

export function imageBuilderVersion(version?: string): string {
  return version || profile.imageBuilderVersion || "2024.10";
}
