import { Observable, BehaviorSubject, from, isObservable, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { switchMap, catchError, map, distinctUntilChanged } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isEqual, flatten, compact } from 'lodash'
import { isPromise, isSubscribable } from '../../util'

/**
 * A service that manages and queries registered notification providers
 * ({@link sourcegraph.NotificationProvider}).
 */
export interface NotificationService {
    /**
     * Observe the notifications provided by registered providers or by a specific provider.
     *
     * @param type Only observe notifications from the provider registered with this type. If
     * undefined, notifications from all providers are observed.
     */
    observeNotifications(
        scope: Parameters<sourcegraph.NotificationProvider['provideNotifications']>[0],
        type?: Parameters<typeof sourcegraph.notifications.registerNotificationProvider>[0]
    ): Observable<sourcegraph.Notification[]>

    /**
     * Register a notification provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerNotificationProvider: typeof sourcegraph.notifications.registerNotificationProvider
}

/**
 * Creates a new {@link NotificationService}.
 */
export function createNotificationService(logErrors = true): NotificationService {
    interface Registration {
        type: Parameters<typeof sourcegraph.notifications.registerNotificationProvider>[0]
        provider: sourcegraph.NotificationProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])
    return {
        observeNotifications: (scope, type) => {
            return registrations.pipe(
                switchMap(registrations =>
                    combineLatestOrDefault(
                        (type === undefined ? registrations : registrations.filter(r => r.type === type)).map(
                            ({ provider }) =>
                                fromProviderResult(provider.provideNotifications(scope), items => items || []).pipe(
                                    catchError(err => {
                                        if (logErrors) {
                                            console.error(err)
                                        }
                                        return [null]
                                    })
                                )
                        )
                    ).pipe(
                        map(itemsArrays => flatten(compact(itemsArrays))),
                        distinctUntilChanged((a, b) => isEqual(a, b))
                    )
                )
            )
        },
        registerNotificationProvider: (type, provider) => {
            if (registrations.value.some(r => r.type === type)) {
                throw new Error(`a NotificationProvider of type ${JSON.stringify(type)} is already registered`)
            }
            const reg: Registration = { type, provider }
            registrations.next([...registrations.value, reg])
            const unregister = () => registrations.next(registrations.value.filter(r => r !== reg))
            return { unsubscribe: unregister }
        },
    }
}

/**
 * Returns an {@link Observable} that represents the same result as a
 * {@link sourcegraph.ProviderResult}, with a mapping.
 *
 * @param result The result returned by the provider
 * @param mapFunc A function to map the result into a type that does not (necessarily) include `|
 * undefined | null`.
 */
function fromProviderResult<T, R>(
    result: sourcegraph.ProviderResult<T>,
    mapFunc: (value: T | undefined | null) => R
): Observable<R> {
    let observable: Observable<R>
    if (result && (isPromise(result) || isObservable<T>(result) || isSubscribable(result))) {
        observable = from(result).pipe(map(mapFunc))
    } else {
        observable = of(mapFunc(result))
    }
    return observable
}