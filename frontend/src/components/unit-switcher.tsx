"use client";

import { SelectField } from "@cia-da-vacina/design-system";
import { useAuth } from "@/contexts/auth-context";

export function UnitSwitcher() {
  const { units, activeUnitId, setUnit } = useAuth();

  if (units.length === 0) return null;

  return (
    <SelectField
      aria-label="Unidade"
      fieldSize="sm"
      fullWidth={false}
      appearance="quiet"
      value={activeUnitId ?? ""}
      onChange={(e) => setUnit(e.target.value)}
      options={units.map((u) => ({ value: u.id, label: u.name }))}
    />
  );
}
