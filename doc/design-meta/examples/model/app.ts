import { UUID, Version, Command } from "./common";
import { RepositoryRef, ConfigSource, DistributionRule } from "./distributed-config-cli";

export type Application = {
  id: UUID;
  name: string;
  title: string;
  version: Version;
};

export type CliApplication = Application & {
  command: Command;
  repositories: RepositoryRef[];
  sources: ConfigSource[];
  rules: DistributionRule[];
};
