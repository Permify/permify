import AWS from 'aws-sdk'

AWS.config.update({
    accessKeyId: process.env.ACCESS_KEY,
    secretAccessKey: process.env.SECRET_KEY
});

const storage = new AWS.S3({
    params: {Bucket: process.env.S3_BUCKET},
    region: process.env.REGION,
})

export default function Upload(file) {
    const params = {
        Body: file,
        Bucket: process.env.S3_BUCKET,
        Key: file.name,
        ContentType: file.type,
    };

    return storage.upload(params).promise();
}
