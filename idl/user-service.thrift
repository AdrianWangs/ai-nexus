// user_service.thrift
namespace go user_microservice

// 定义用户信息结构体
struct User {
    1: i64 UserId,
    2: string Username,
    3: string Password,
    4: string Birthday,
    5: string Gender,
    6: i32 RoleId,
    7: string PhoneNumber,
    8: string Email,
    9: optional string ThirdPartyToken, // 第三方登录token，可选
}

// 登录请求结构体
struct LoginRequest {
    1: string UsernameOrEmail,
    2: string Password, // 密码或第三方token
}

// 登录响应结构体
struct LoginResponse {
    1: bool Success,
    2: optional string ErrorMessage,
    3: optional User UserProfile,
    4: optional string Token,
}

// 注册请求结构体
struct RegisterRequest {
    1: string Username,
    2: string Password,
    3: string Email,
    4: string PhoneNumber,
    5: string Birthday,
    6: string Gender,
}

// 注册响应结构体
struct RegisterResponse {
    1: bool Success,
    2: optional string ErrorMessage,
}

// 第三方登录请求结构体，假设通过OAuth等
struct ThirdPartyLoginRequest {
    1: string Token,
    2: string Provider, // 如："facebook", "google"
}

// 第三方登录响应结构体
struct ThirdPartyLoginResponse {
    1: bool Success,
    2: optional string ErrorMessage,
    3: optional User UserProfile,
    4: optional string Token,
}

// 修改用户信息请求结构体
struct UpdateUserRequest {
    1: string Username,
    2: optional string Password,
    3: optional string Email,
    4: optional string PhoneNumber,
    5: optional string Birthday,
    6: optional string Gender,
    7: optional string ThirdPartyToken,
}

// 修改用户信息响应结构体
struct UpdateUserResponse {
    1: bool Success,
    2: optional string ErrorMessage,
}

struct GetUserRequest {
    1: i64 UserId,
}

struct GetUserResponse {
    1: bool Success,
    2: optional string ErrorMessage,
    3: optional User UserProfile,
}

// 定义用户服务接口
service UserService {
    // 用户登录
    LoginResponse Login(1: LoginRequest request)(api.post="login");

    // 用户注册
    RegisterResponse Register(1: RegisterRequest request)(api.post="register");

    // 第三方登录
    ThirdPartyLoginResponse ThirdPartyLogin(1: ThirdPartyLoginRequest request)(api.post="third_party_login");

    // 修改用户信息
    UpdateUserResponse UpdateUserProfile(1: UpdateUserRequest request)(api.post="update_user_profile");

    // 获取用户信息
    GetUserResponse GetUser(1: GetUserRequest request)(api.get="get_user");

}(
    api.path="/user",
    api.version="v1",
    api.description="用户服务"
)