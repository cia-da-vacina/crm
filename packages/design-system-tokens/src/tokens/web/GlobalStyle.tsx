import { createGlobalStyle } from "styled-components";
import raw from "../global/colors";

const GlobalStyle = createGlobalStyle`
  *, *::before, *::after {
    box-sizing: border-box;
  }

  html, body {
    margin: 0;
    padding: 0;
    min-height: 100%;
  }

  body {
    font-family: ${({ theme }) => theme?.fonts?.body ?? '"DM Sans", sans-serif'};
    font-size: ${({ theme }) => theme?.fontSizes?.body ?? "14px"};
    line-height: ${({ theme }) => theme?.lineHeights?.normal ?? 1.5};
    color: ${({ theme }) => theme?.colors?.["text.primary"] ?? "#1B2420"};
    background:
      radial-gradient(1100px 480px at 8% -12%, ${raw.evergreen[50]}, transparent 55%),
      radial-gradient(900px 420px at 92% 0%, ${raw.mist[50]}, transparent 50%),
      ${({ theme }) => theme?.colors?.["bg.canvas"] ?? raw.mist[25]};
    -webkit-font-smoothing: antialiased;
  }

  a {
    color: ${({ theme }) => theme?.colors?.["text.link"] ?? raw.evergreen[600]};
    text-decoration: none;
  }

  button, input, select, textarea {
    font: inherit;
  }

  ::selection {
    background: ${raw.evergreen[100]};
    color: ${raw.evergreen[900]};
  }
`;

export default GlobalStyle;
