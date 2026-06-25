// SPDX-License-Identifier: Apache-2.0
//
// Small helpers for parsing recipient input fields.

/** Split a comma/semicolon-separated address field and extract bare addresses. */
export function parseAddressList(input: string): string[] {
  return input
    .split(/[,;\n]/)
    .map((part) => extractAddress(part))
    .filter((addr): addr is string => !!addr);
}

/** Pull the email out of "Name <addr@host>" or a bare "addr@host". */
export function extractAddress(value: string): string | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  const angle = trimmed.match(/<([^>]+)>/);
  const candidate = angle ? angle[1].trim() : trimmed;
  return /.+@.+\..+/.test(candidate) ? candidate : null;
}
