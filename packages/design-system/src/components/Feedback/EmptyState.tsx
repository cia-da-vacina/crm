import Stack from "../Layout/Stack";
import Heading from "../Typography/Heading";
import Text from "../Typography/Text";
import type { ReactNode } from "react";

export type EmptyStateProps = {
  title: string;
  description?: string;
  icon?: ReactNode;
  action?: ReactNode;
};

export default function EmptyState({
  title,
  description,
  icon,
  action,
}: EmptyStateProps) {
  return (
    <Stack alignItems="center" justifyContent="center" py={9} gap={3} textAlign="center">
      {icon}
      <Heading as="h3">{title}</Heading>
      {description && <Text muted>{description}</Text>}
      {action}
    </Stack>
  );
}
