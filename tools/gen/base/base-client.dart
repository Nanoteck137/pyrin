import 'dart:convert';

import 'package:result_dart/result_dart.dart';
import 'package:dio/dio.dart';

class ApiError {
  const ApiError(this.type, this.code, this.message);

  final String type;
  final int code;
  final String message;
}

class NoBody {}

class RequestOptions {
  Map<String, dynamic>? query;

  RequestOptions({this.query});
}

class BaseApiClient {
  BaseApiClient({String baseUrl = ""}) {
    _dio = Dio(
      BaseOptions(
        baseUrl: baseUrl,
        validateStatus: (status) {
          return true;
        },
      ),
    );
  }

  late Dio _dio;

  String? _accessToken;
  String? _apiToken;

  void setAccessToken(String? token) {
    _accessToken = token;
  }

  void setApiToken(String? token) {
    _apiToken = token;
  }

  AsyncResultDart<Map<String, dynamic>, ApiError> request(
    String method,
    String path, {
    RequestOptions? options,
    Map<String, dynamic>? body,
  }) async {
    final headers = <String, dynamic>{};

    if (_accessToken != null) {
      headers["Authorization"] = "Bearer $_accessToken";
    }

    if (_apiToken != null) {
      headers["X-Api-Token"] = _apiToken;
    }

    final res = await _dio.request(
      path,
      options: Options(method: method, headers: headers),
      queryParameters: options?.query,
      data: body != null ? jsonEncode(body) : null,
    );

    final data = res.data as Map<String, dynamic>;

    final success = data["success"] as bool;

    if (!success) {
      final error = data["error"] as Map<String, dynamic>;
      return Failure(
        ApiError(
          error["type"] as String,
          error["code"] as int,
          error["message"] as String,
        ),
      );
    }

    return Success(data["data"]);
  }
}
