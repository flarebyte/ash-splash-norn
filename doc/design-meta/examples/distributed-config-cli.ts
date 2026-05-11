import { UUID, Version } from "./common";

export type ConfigFormat = "cue";
export type OutputFormat = "json" | "yaml" | "go" | "dart";

export type RepositoryRef = {
  id: UUID;
  name: string;
  rootPath: string;
  branch?: string;
};

export type ConfigSource = {
  id: UUID;
  repositoryId: UUID;
  format: ConfigFormat;
  path: string;
};

export type OutputTarget = {
  format: OutputFormat;
  path: string;
  packageName?: string;
};

export type DistributionRule = {
  id: UUID;
  sourceId: UUID;
  targetRepositories: UUID[];
  outputs: OutputTarget[];
};

export type DistributedConfigCliApp = {
  id: UUID;
  name: string;
  version: Version;
  repositories: RepositoryRef[];
  sources: ConfigSource[];
  rules: DistributionRule[];
};
