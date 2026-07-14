export { default as CiaThemeProvider } from "./provider/CiaThemeProvider";

export { default as Box } from "./components/Layout/Box";
export type { BoxProps } from "./components/Layout/Box";
export { default as Flex } from "./components/Layout/Flex";
export type { FlexProps } from "./components/Layout/Flex";
export { default as Stack } from "./components/Layout/Stack";
export type { StackProps } from "./components/Layout/Stack";

export { default as Text } from "./components/Typography/Text";
export type { TextProps } from "./components/Typography/Text";
export { default as Heading } from "./components/Typography/Heading";
export type { HeadingProps } from "./components/Typography/Heading";

export { default as Button } from "./components/Input/Button";
export type { ButtonProps } from "./components/Input/Button";
export { default as TextField } from "./components/Input/TextField";
export type { TextFieldProps } from "./components/Input/TextField";
export { default as SelectField } from "./components/Input/SelectField";
export type { SelectFieldProps, SelectOption, SelectFieldSize } from "./components/Input/SelectField";

export { Badge, StageBadge } from "./components/DataDisplay/Badge";
export type { BadgeProps, BadgeTone, StageBadgeProps } from "./components/DataDisplay/Badge";
export { default as Avatar } from "./components/DataDisplay/Avatar";
export type { AvatarProps } from "./components/DataDisplay/Avatar";

export { default as EmptyState } from "./components/Feedback/EmptyState";
export { default as Spinner } from "./components/Feedback/Spinner";

export { default as AppShell } from "./components/Navigation/AppShell";
export type { AppShellProps, AppShellLink } from "./components/Navigation/AppShell";

export { default as ConversationRow } from "./components/CRM/ConversationRow";
export type { ConversationRowProps } from "./components/CRM/ConversationRow";
export { default as ConversationList } from "./components/CRM/ConversationList";

export {
  webLight,
  stageLabels,
  pipelineStages,
} from "@cia-da-vacina/design-system-tokens";
export type { PipelineStage } from "@cia-da-vacina/design-system-tokens";
