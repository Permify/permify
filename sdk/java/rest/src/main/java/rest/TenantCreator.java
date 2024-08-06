package rest;

import org.joda.time.LocalTime;
import reactor.core.publisher.Mono;

import org.permify.ApiClient;
import org.permify.api.TenancyApi;

import org.permify.model.TenantCreateRequest;
import org.permify.model.TenantCreateResponse;

public class TenantCreator {
  public static void main(String[] args) {
    String baseUrl = "http://localhost:3476"; // Ensure this is the correct port
    String bearerToken = "secret";  // Replace with your actual bearer token

    ApiClient apiClient = new ApiClient();
    apiClient.setBasePath(baseUrl);
    System.out.println("Created Api Client Endpoint: " + apiClient.getBasePath());
    apiClient.addDefaultHeader("Authorization", "Bearer secret");

    TenancyApi api = new TenancyApi(apiClient);

    try {
        System.out.println("Sending request to " + baseUrl + " with Authorization header Bearer " + bearerToken);
        TenantCreateRequest req = new TenantCreateRequest();
        req.setId("template_tenant2");
        req.setName("Template");
        Mono<TenantCreateResponse> responseMono = api.tenantsCreate(req);
        TenantCreateResponse response = responseMono.block(); // Block to get the response synchronously
        System.out.println("Tenant create response: " + response);
    } catch (Exception e) {
        LocalTime time = new LocalTime();
        System.out.println("Error occurred at " + time.toString() + ": " + e.getMessage());
    }
  }
}
