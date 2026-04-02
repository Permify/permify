import {processBlobUpload} from '../../blob-upload.js';

export default async function handler(request) {
  const body = await request.json();

  try {
    const jsonResponse = await processBlobUpload({request, body});
    return Response.json(jsonResponse);
  } catch (error) {
    return Response.json(
      {error: error instanceof Error ? error.message : 'Blob upload failed.'},
      {status: 400}
    );
  }
}
