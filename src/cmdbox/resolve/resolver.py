from typing import Optional

from .errors import MaxDepthExceeded, UnknownReference, CycleDetectionError
from .lookup import ResolverLookup
from .type_defs import ResolveResult, TraceStep, RefKind


class Resolver:
    """
    Resolves templates by substituting tokens enclosed in angular brackets (<...>) with their
    corresponding values fetched from a predefined lookup. Supports variable and command
    resolution, recursive expansion, and cycle detection.

    This class is designed to process textual templates with customizable behavior. It provides
    support for tracing resolution steps, enforcing recursion depth limits, and handling both
    commands and variables with a flexible lookup mechanism.

    Attributes:
        lookup (ResolverLookup): A lookup mechanism for resolving commands and variables.
        strict (bool): Enforces strict error handling. If True, raises errors for unresolved
            tokens; otherwise, leaves unresolved tokens in their original form.
        max_depth (int): The allowed maximum depth for nested token substitutions. Prevents
            excessively deep or recursive expansions. Defaults to 25.
    """

    def __init__(
        self, lookup: ResolverLookup, *, strict: bool = False, max_depth: int = 25
    ):
        self._lookup = lookup
        self._strict = strict
        self._max_depth = max_depth

    def resolve(
        self,
        template: str,
        root_label: str = "<input>",
        runtime_vars: dict[str, str] | None = None,
    ) -> ResolveResult:
        """
        Resolves a given template string by iterating through its elements and performing
        text and stack manipulations until a final result is obtained. This method employs
        recursive resolution using internal helper functions and produces a structured
        output containing the resolved text and trace steps.

        Args:
            template (str): The template string to resolve.
            root_label (str): The initial label to use as the root for resolution. Defaults
                to "<input>".
            runtime_vars (dict[str, str] | None): Variables supplied by the user at runtime
                to be used during resolution. Defaults to None.

        Returns:
            ResolveResult: The result of the resolution, including the resolved text
            and the list of trace steps.
        """
        trace: list[TraceStep] = []
        stack: list[str] = [root_label]
        text = self._resolve_inner(
            template,
            stack=stack,
            depth=0,
            trace=trace,
            runtime_vars=runtime_vars or {},
        )
        return ResolveResult(text=text, trace=trace)

    def collect_missing_vars(
        self, template: str, runtime_vars: dict[str, str] | None = None
    ) -> list[str]:
        """
        Collects missing variables from the given template.

        This function analyzes the provided template and checks for any placeholders
        or variables that are not present in the runtime variables provided. It
        returns a list of all such missing variables that need to be defined or
        replaced in the template.

        Args:
            template (str): The template string containing placeholder variables
                to be checked.
            runtime_vars (dict[str, str] | None): A dictionary of available
                runtime variables where keys are variable names and values are
                their corresponding values. Defaults to None if not provided.

        Returns:
            list[str]: A list of variable names that are missing in the provided
                runtime variables.
        """
        missing: list[str] = []
        self._collect_missing_inner(
            template,
            runtime_vars=runtime_vars or {},
            missing=missing,
            seen=set(),
        )
        return missing

    def _resolve_inner(
        self,
        template: str,
        *,
        stack: list[str],
        depth: int,
        trace: list[TraceStep],
        runtime_vars: dict[str, str] | None = None,
    ) -> str:
        """
        Resolves the given template string with nested tokens, handling escapes and
        recursively processing angle-bracketed tokens until the maximum depth or the
        end of the template is reached.

        Args:
            template: The template string containing tokens to resolve.
            stack: The stack of currently resolved tokens to avoid circular references.
            depth: The current recursion depth, used to track and prevent exceeding the
                maximum allowed depth.
            trace: A list of TraceStep objects for debugging and understanding the
                resolution flow.

        Returns:
            A resolved string with tokens replaced by their corresponding values or
            content.

        Raises:
            MaxDepthExceeded: When the recursion depth exceeds the allowed maximum.
        """
        if depth > self._max_depth:
            raise MaxDepthExceeded(self._max_depth)

        out: list[str] = []
        i = 0  # Current index
        t_len = len(template)

        while i < t_len:
            ch = template[i]
            if ch == "\\":
                if i + 1 < t_len and template[i + 1] in ("\\", "<", ">"):
                    out.append(template[i + 1])
                    i += 2
                    continue

            if ch == "<":
                token_inner, raw_token, next_i = self._read_angle_token(template, i)
                if token_inner is None:
                    out.append("<")
                    i += 1
                    continue
                replacement = self._expand_angle_token(
                    token_inner,
                    raw_token=raw_token,
                    stack=stack,
                    depth=depth,
                    trace=trace,
                    runtime_vars=runtime_vars,
                )
                out.append(replacement)
                i = next_i
                continue
            out.append(ch)
            i += 1

        return "".join(out)

    def _collect_missing_inner(
        self,
        template: str,
        *,
        runtime_vars: dict[str, str],
        missing: list[str],
        seen: set[str],
        depth: int = 0,
    ) -> None:
        """
        Recursively collects missing variables and commands in the template by analyzing
        placeholders and their references. It ensures that variable and command dependencies
        are resolved to either runtime variables or previously defined variables or commands.
        If a placeholder cannot be resolved within the specified depth, it is added to the
        missing list for further processing.

        Args:
            template: The template string containing placeholders to analyze.
            runtime_vars: A dictionary mapping variable names to their values. Runtime
                variables are used to satisfy some of the placeholders in the template.
            missing: A list used to collect the names of unresolved variable placeholders.
            seen: A set used to track the names of variables and commands that have already
                been processed, avoiding circular dependencies during recursion.
            depth: The current recursion depth. Defaults to 0. If the recursion depth exceeds
                the allowed maximum (self._max_depth), a MaxDepthExceeded exception is raised.
        """
        if depth > self._max_depth:
            return
        i = 0
        while i < len(template):
            if template[i] == "\\":
                i += 2
                continue
            if template[i] == "<":
                token_inner, _, next_i = self._read_angle_token(template, i)
                if token_inner:
                    kind, key = self._parse_kind_and_key(token_inner)
                    if kind == RefKind.VARIABLE:
                        if runtime_vars and key in runtime_vars:
                            pass  # satisfied
                        elif self._lookup.get_variable(key) is None:
                            if key not in missing:
                                missing.append(key)
                        else:
                            # Recurse into stored variable's value
                            rec = self._lookup.get_variable(key)
                            if rec and rec.name not in seen:
                                seen.add(rec.name)
                                self._collect_missing_inner(
                                    rec.value,
                                    runtime_vars=runtime_vars,
                                    missing=missing,
                                    seen=seen,
                                    depth=depth + 1,
                                )
                    else:  # COMMAND ref
                        rec = self._lookup.get_command(key)
                        if rec and rec.alias not in seen:
                            seen.add(rec.alias)
                            self._collect_missing_inner(
                                rec.template,
                                runtime_vars=runtime_vars,
                                missing=missing,
                                seen=seen,
                                depth=depth + 1,
                            )
                i = next_i
            else:
                i += 1

    def _read_angle_token(self, s: str, start_i: int) -> tuple[Optional[str], str, int]:
        """
        Reads a token beginning with "<" until an unescaped ">" is encountered.

        Args:
            s (str): The string containing the token.
            start_i (int): The index of the first character of the token.

        Returns:
            str: token_inner (with escapes already interpreted), or None if no closing bracket.
            str: raw_token - the original token string, including the angle brackets.
            int: next_i - the index of the first character after the closing bracket.

        """
        assert s[start_i] == "<"
        i = start_i + 1
        n = len(s)

        inner_chars: list[str] = []

        while i < n:
            ch = s[i]
            if ch == "\\":
                if i + 1 < n and s[i + 1] in ("\\", "<", ">"):
                    inner_chars.append(s[i + 1])
                    i += 2
                    continue
                inner_chars.append("\\")
                i += 1
                continue

            if ch == ">":
                raw_token = s[start_i + 1 : i]
                token_inner = "".join(inner_chars).strip()
                return token_inner, raw_token, i + 1

            inner_chars.append(ch)
            i += 1

        return None, s[start_i:], n

    def _expand_angle_token(
        self,
        token_inner: str,
        *,
        raw_token: str,
        stack: list[str],
        depth: int,
        trace: list[TraceStep],
        runtime_vars: dict[str, str] | None = None,
    ) -> str:
        """
        Expands a token enclosed within angular brackets (<...>) into its resolved value
        based on predefined rules such as matching to commands or variables. It uses an
        internal lookup mechanism and checks for cyclic references during the resolution
        process.

        Args:
            token_inner (str): The content within angular brackets to be resolved.
            raw_token (str): The original token as a fallback if resolution fails.
            stack (list[str]): A stack used for detecting cyclic dependencies during token
                resolution.
            depth (int): The current depth of recursion for token resolution. This is used
                to track and limit nesting levels.
            trace (list[TraceStep]): A list of TraceStep objects that document the steps
                during the resolution of the token.

        Returns:
            str: The expanded and resolved token string.
        """
        if not token_inner:
            return raw_token
        kind, key = self._parse_kind_and_key(token_inner)

        if kind == RefKind.VARIABLE:

            # Runtime vars override DB vars, check for them first
            if runtime_vars and key in runtime_vars:
                value = runtime_vars[key]
                trace.append(
                    TraceStep(kind=RefKind.VARIABLE, key=key, expanded_to=value)
                )
                return value

            rec = self._lookup.get_variable(key)
            if rec is None:
                if self._strict:
                    raise UnknownReference("variable", key)
                return raw_token

            label = f"var:{rec.name}"
            self._check_cycle(label, stack)
            stack.append(label)
            try:
                expanded = self._resolve_inner(
                    rec.value, stack=stack, depth=depth + 1, trace=trace
                )
            finally:
                stack.pop()

            trace.append(
                TraceStep(kind=RefKind.VARIABLE, key=rec.name, expanded_to=expanded)
            )
            return expanded

        # Fallback to command lookup
        rec = self._lookup.get_command(key)
        if rec is None:
            if self._strict:
                raise UnknownReference("command", key)
            return raw_token

        label = f"cmd:{rec.alias}"
        self._check_cycle(label, stack)
        stack.append(label)
        try:
            expanded = self._resolve_inner(
                rec.template, stack=stack, depth=depth + 1, trace=trace
            )
        finally:
            stack.pop()
        trace.append(
            TraceStep(kind=RefKind.COMMAND, key=rec.alias, expanded_to=expanded)
        )
        return expanded

    def _parse_kind_and_key(self, token_inner: str) -> tuple[RefKind, str]:
        """
        Parses the kind and key from a given token string.

        This method evaluates the provided token and determines its type, which can be a
        command or variable, based on the prefix present in the token. If no prefix is
        found, it defaults to `RefKind.VARIABLE`. The method ensures that parts of the
        token are appropriately stripped of whitespace during processing.

        Args:
            token_inner (str): The input token string that may specify a prefix and a key,
                separated by a colon.

        Returns:
            tuple[RefKind, str]: A tuple containing the kind of reference (`RefKind`) and
            its corresponding key as a string.
        """
        if ":" not in token_inner:
            return RefKind.VARIABLE, token_inner
        prefix, key = (part.strip() for part in token_inner.split(":", 1))

        if prefix == "cmd":
            return RefKind.COMMAND, key
        if prefix == "var":
            return RefKind.VARIABLE, key
        return RefKind.VARIABLE, token_inner

    def _check_cycle(self, next_label: str, stack: list[str]) -> None:
        """
        Checks for a cycle in the provided stack with respect to the `next_label`.

        This method checks whether the `next_label` appears in the current stack,
        indicating the presence of a cycle. If a cycle is detected, a
        `CycleDetectionError` is raised.

        Args:
            next_label (str): The label to check within the stack to detect a cycle.
            stack (list[str]): List maintaining the current sequence of elements being
                evaluated.
        """
        if next_label in stack:
            cycle_start = stack.index(next_label)
            raise CycleDetectionError(stack[cycle_start:] + [next_label])
