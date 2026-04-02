export const getBaseUrl = () => import.meta.env.BASE_URL || '/';

export const getBlobToken = () =>
  import.meta.env.VITE_BLOB_READ_WRITE_TOKEN ||
  import.meta.env.REACT_APP_BLOB_READ_WRITE_TOKEN ||
  '';
