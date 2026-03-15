## 2026-03-15 - Accessible Icon Buttons in Social Feed
**Learning:** Adding `aria-label` to icon-only buttons is crucial, but for buttons with visible text (e.g., "Like 2.4K"), overriding the entire label with `aria-label` hides the numerical count from screen readers.
**Action:** Use `.visually-hidden` span elements alongside the visible counts instead of `aria-label` for buttons that contain dynamic text counts.
