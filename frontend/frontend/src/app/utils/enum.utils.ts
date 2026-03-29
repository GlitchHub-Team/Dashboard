export class EnumMapper<TFrontend extends string, TBackend extends string> {
  private readonly toBackendMap: Record<string, string>;
  private readonly toFrontendMap: Record<string, string>;
  private readonly fallback: TFrontend;

  constructor(mapping: Record<TFrontend, TBackend>, fallback: TFrontend) {
    this.fallback = fallback;
    this.toBackendMap = { ...mapping };
    this.toFrontendMap = Object.entries(mapping).reduce(
      (acc, [frontend, backend]) => {
        acc[backend as string] = frontend;
        return acc;
      },
      {} as Record<string, string>,
    );
  }

  toBackend(value: TFrontend): TBackend {
    const mapped = this.toBackendMap[value];
    return mapped as TBackend;
  }

  fromBackend(value: string): TFrontend {
    const mapped = this.toFrontendMap[value];
    if (!mapped) {
      return this.fallback;
    }
    return mapped as TFrontend;
  }
}
