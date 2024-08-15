package grpc;

import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

import build.buf.gen.base.v1.TenancyGrpc;
import build.buf.gen.base.v1.TenantCreateRequest;
import build.buf.gen.base.v1.TenantCreateResponse;


public class TenantCreator {
  public static void main(String[] args) {
      // Create the channel
      ManagedChannel channel = ManagedChannelBuilder.forAddress("127.0.0.1", 3478).usePlaintext().build();

      try {
          // Create the blocking stub
          TenancyGrpc.TenancyBlockingStub blockingStub = TenancyGrpc.newBlockingStub(channel);

          String timeStamp = new SimpleDateFormat("yyyy.MM.dd.HH.mm.ss").format(new Date());
          TenantCreateRequest req = TenantCreateRequest.newBuilder()
                  .setId("tenant_" + timeStamp)
                  .setName("tenant id name")
                  .build();

          TenantCreateResponse response = blockingStub.create(req);
          System.out.println(response);

      } finally {
          // Gracefully shutdown the channel
          try {
              channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
          } catch (InterruptedException e) {
              e.printStackTrace();
              channel.shutdownNow();
          }
      }
  }
}