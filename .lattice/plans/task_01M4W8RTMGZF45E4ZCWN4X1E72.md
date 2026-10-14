# BC-21: FilterBar: announce invalid-submit and server-error feedback to screen readers

From BC-9 code review (art_01M4W8RJEX9PZGNYXK1279SVG3, nit 5): messages revealed by an invalid Enter or a server 400 are linked via aria-describedby but are not in a live region, so a screen-reader user who presses Enter on an invalid bar gets no feedback. SPEC §7.6 does not require it (out of BC-9 scope). Candidate fix: on an invalid submit, move focus to the first invalid field (its aria-describedby message is then read); or a polite live summary. Keep BRANDING §5.2 visuals unchanged.
