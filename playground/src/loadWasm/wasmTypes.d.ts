declare global {
  export interface Window {
    Go: any;
    run: (shape: string)=>any[]
    visualize: ()=>any[]
  }
}

export {};
