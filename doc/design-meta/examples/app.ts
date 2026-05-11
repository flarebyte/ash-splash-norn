import { UUID, Version, Command } from "./common";
import { i18nLabelKey } from "./i18n";
import { RepositoryRef, ConfigSource, DistributionRule } from "./distributed-config-cli";

export type Application = {
  id: UUID;
  name: i18nLabelKey;
  title: i18nLabelKey;
  version: Version;
};

export type CliApplication = Application & {
  command: Command;
  repositories: RepositoryRef[];
  sources: ConfigSource[];
  rules: DistributionRule[];
};
