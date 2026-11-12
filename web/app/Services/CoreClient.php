<?php
namespace App\Services;
use Illuminate\Support\Facades\Http;
use RuntimeException;
class CoreClient { public function request(string $method,string $path,array $query=[],?array $body=null): array {
  foreach ($query as $key => $value) { if (is_bool($value)) $query[$key] = $value ? 'true' : 'false'; }
  $request=Http::withToken(config('services.core.token'))->acceptJson()->timeout(25);
  $url=rtrim(config('services.core.url'),'/').$path;
  $response=$request->send(strtoupper($method),$url,['query'=>$query,'json'=>$body]);
  $payload=$response->json();
  if(!$response->successful()||!is_array($payload)||!($payload['ok']??false)){
   $message=is_array($payload)?($payload['error']['message']??'mx-core request failed'):'mx-core returned invalid response';
   throw new RuntimeException($message);
  }
  return is_array($payload['data']??null)?$payload['data']:[];
 }
 public function getServerVersion(): array { $response=Http::withToken(config('services.core.token'))->acceptJson()->timeout(12)->get(rtrim(config('services.core.url'),'/').'/api/v1/server/version');if(!$response->successful()){throw new RuntimeException('mx-core returned HTTP '.$response->status());}$body=$response->json();if(!is_array($body)||!($body['ok']??false)){throw new RuntimeException('mx-core returned an invalid response');}return $body['data']??[]; } }
