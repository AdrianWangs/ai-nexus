// 旅游相关的服务，可以查找景点、查询票价等旅游相关的功能
namespace go nexus_microservice

// 旅游景点信息
struct TouristSpot {
    1: string id,
    2: string name,
    3: string description,
    4: string location,
    5: list<string> images
}

// 旅游计划
struct TravelPlan {
    1: string id,
    2: string title,
    3: string description,
    4: list<TouristSpot> spots,
    5: i32 durationDays
}

// 制定旅游计划请求
struct CreateTravelPlanRequest {
    1: TravelPlan plan,
}

// 制定旅游计划响应
struct CreateTravelPlanResponse {
    1: bool success,
    2: string message,
    3: TravelPlan plan,
}

// 执行旅游计划请求（可以是标记计划开始、完成等动作）
struct ExecuteTravelPlanRequest {
    1: string planId,
    2: string action, // 如 "start", "complete" 等
}

// 执行旅游计划响应
struct ExecuteTravelPlanResponse {
    1: bool success,
    2: string message,
    3: TravelPlan updatedPlan,
}

// 查询旅游景点请求
struct QueryTouristSpotRequest {
    1: string spotId, // 可选，根据ID查询
    2: string nameKeyword, // 可选，根据名称关键字查询
}

// 查询旅游景点响应
struct QueryTouristSpotResponse {
    1: bool success,
    2: string message,
    3: list<TouristSpot> spots,
}

// 旅游景点票价信息
struct TicketPrice {
    1: string spotId,
    2: i32 adultPrice,
    3: i32 childPrice,
    4: string currency,
    5: optional string note, // 可选的备注信息，比如学生票、团体票说明等
}

// 查询旅游景点票价请求
struct QueryTicketPriceRequest {
    1: string spotId,
}

// 查询旅游景点票价响应
struct QueryTicketPriceResponse {
    1: bool success,
    2: string message,
    3: optional TicketPrice price, // 如果没有查到价格，则此字段为null
}

// 旅游服务接口
service TravelPlanService {

    // 查询旅游景点票价
    QueryTicketPriceResponse queryTicketPrice(1: QueryTicketPriceRequest request),
    // 查询旅游景点
    QueryTouristSpotResponse queryTouristSpot(1: QueryTouristSpotRequest request),
}