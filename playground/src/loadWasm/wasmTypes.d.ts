declare global {
  export interface Window {
    Go: any;
    check: (query: string)=>any[]
    lookupEntity: (query: string)=>any[]
    lookupSubject: (query: string)=>any[]
    writeSchema: (schema: string)=>any[]
    readSchema: ()=>any[]
    writeTuple: (tuple: string)=>any[]
    readTuple: (filter: string)=>any[]
    deleteTuple: (tuple: string)=>any[]
    readSchemaGraph: ()=>any[]
  }
}

export {};
