<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class CheckGatewaySecret
{
    /**
     * Handle an incoming request.
     *
     * @param  \Closure(\Illuminate\Http\Request): (\Symfony\Component\HttpFoundation\Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        if($request->header('X-Gateway-Secret') !== env('GATEWAY_SECRET')) {
            return response()->json([
                'status' => 'Error',
                'message' => 'Forbidden'
                ], 403);
        }

        return $next($request);
    }
}
