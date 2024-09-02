namespace go ScheduleService

// 定义基础的数据类型
struct Event {
    1: i64 id,
    2: string title,
    3: string description,
    4: string location,
    5: string startTime,
    6: string endTime,
}

enum EventType {
    MEETING = 1,
    TASK = 2,
    APPOINTMENT = 3,
}

// Event 可能会有不同的类型，这里通过枚举扩展
struct TypedEvent {
    1: EventType eventType,
    2: Event baseEvent,
}

// 服务接口定义
service ScheduleService {
    // 创建一个新的事件
    i64 createEvent(1: Event event),

    // 更新已存在的事件
    bool updateEvent(1: i64 eventId, 2: Event updatedEvent),

    // 删除事件
    bool deleteEvent(1: i64 eventId),

    // 获取特定ID的事件详情
    TypedEvent getEventById(1: i64 eventId),

    // 列出所有事件
    list<TypedEvent> listAllEvents(),
}