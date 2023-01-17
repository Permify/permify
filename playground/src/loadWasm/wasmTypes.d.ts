declare global {
  export interface Window {
    Go: any;
    check: (query: string, version: string)=>any[]
    lookupEntity: (query: string, version: string)=>any[]
    writeSchema: (schema: string)=>any[]
    readSchema: (version: string)=>any[]
    writeTuple: (tuple: string, version: string)=>any[]
    readTuple: (filter: string)=>any[]
    deleteTuple: (tuple: string)=>any[]
    readSchemaGraph: (version: string)=>any[]
  }
}

export {};
