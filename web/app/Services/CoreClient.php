<?php
namespace App\Services;
use Illuminate\Support\Facades\Http;
use RuntimeException;
class CoreClient { public function getServerVersion(): array { $response=Http::withToken(config('services.core.token'))->acceptJson()->timeout(12)->get(rtrim(config('services.core.url'),'/').'/api/v1/server/version');if(!$response->successful()){throw new RuntimeException('mx-core returned HTTP '.$response->status());}$body=$response->json();if(!is_array($body)||!($body['ok']??false)){throw new RuntimeException('mx-core returned an invalid response');}return $body['data']??[]; } }
