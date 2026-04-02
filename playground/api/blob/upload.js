import {processBlobUpload} from '../../blob-upload.js';

async function getRequestBody(request) {
  if (request.body) {
    return request.body;
  }

  if (typeof request.json === 'function') {
    return request.json();
  }

  throw new Error('Unsupported upload request body.');
}

export default async function handler(request, response) {
  try {
    const body = await getRequestBody(request);
    const jsonResponse = await processBlobUpload({request, body});

    if (response) {
      return response.status(200).json(jsonResponse);
    }

    return Response.json(jsonResponse);
  } catch (error) {
    const payload = {error: error instanceof Error ? error.message : 'Blob upload failed.'};

    if (response) {
      return response.status(400).json(payload);
    }

    return Response.json(
      payload,
      {status: 400}
    );
  }
}
