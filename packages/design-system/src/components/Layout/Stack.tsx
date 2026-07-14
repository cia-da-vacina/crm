import Flex, { type FlexProps } from "./Flex";

export type StackProps = FlexProps & {
  direction?: "column" | "row";
};

export default function Stack({
  direction = "column",
  gap = 3,
  ...rest
}: StackProps) {
  return <Flex flexDirection={direction} gap={gap} {...rest} />;
}
