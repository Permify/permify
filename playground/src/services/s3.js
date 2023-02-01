import AWS from 'aws-sdk'

const S3_BUCKET = 'permify.playground.storage';
const REGION = 'us-east-1';

AWS.config.update({
    accessKeyId: '{{ACCESS_KEY}}',
    secretAccessKey: '{{SECRET_KEY}}'
})

const storage = new AWS.S3({
    params: {Bucket: S3_BUCKET},
    region: REGION,
})

export default function Upload(file) {
    const params = {
        Body: file,
        Bucket: S3_BUCKET,
        Key: file.name,
        ContentType: file.type,
    };

    return storage.upload(params).promise();
}
