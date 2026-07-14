import styled from "styled-components";
import Box, { type BoxProps } from "./Box";

export type FlexProps = BoxProps;

const Flex = styled(Box)<FlexProps>`
  display: flex;
`;

export default Flex;
