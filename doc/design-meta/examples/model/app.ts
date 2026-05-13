import { UUID, Version, Command } from "./common";

export type Application = {
  id: UUID;
  name: string;
  title: string;
  version: Version;
};

export type CliApplication = Application & {
  command: Command;
  repositoryIds: UUID[];
  sourceIds: UUID[];
  ruleIds: UUID[];
};
